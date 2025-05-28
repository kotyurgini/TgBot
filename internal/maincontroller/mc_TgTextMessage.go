package maincontroller

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"tgbot/internal/ai"
	"tgbot/internal/localization"
	"tgbot/internal/store"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (mc *MainController) handleTgMessage(req *Request) *MessageManager {
	method := "handleTgMessage()"
	us := req.UserShell
	text := ""
	if req.Update.Message != nil {
		text = req.Update.Message.Text
	} else {
		text = req.UserShell.LastText
	}

	msgEx := newMessageExchange()

	go func() {
		defer msgEx.close()
		_, hasRequest := mc.hasActiveUserRequest(us.ID)
		if hasRequest {
			msg := newTgMessage(us.ID, localeText(us.Locale, localization.MTypeAnswerCreateNewDialog))
			msg.ReplyMarkup = kbWithOneButton(
				"",
				localeText(us.Locale, localization.MTypeBtnCancelPreviousRequest),
				fmt.Sprint(callbackTypeCancelRequest))
			_, _ = msgEx.send(msg)
			return
		}

		checkLimit, err := mc.store.CheckUserUsage(us, us.User.ChatModelID)
		if err != nil {
			msgEx.sendError(err)
			return
		}
		if !checkLimit {
			msg := newTgMessage(us.ID, localeText(us.Locale, localization.MTypeMsgLimitReached))
			_, _ = msgEx.send(msg)
			return
		}

		checkNewDialog, err := CheckLastMessageTime(mc, us, text, msgEx)
		if err != nil {
			msgEx.sendError(err)
			return
		}
		if checkNewDialog {
			return
		}

		if req.UserShell.Dialog == nil {
			_, err := mc.store.AddDialog(req.Ctx, req.UserShell, &store.Dialog{Title: dialogTitle(text), UserID: us.ID, Created: time.Now().UTC()})
			if err != nil {
				msgEx.sendError(fmt.Errorf("%s: %w", method, err))
				return
			}
		}

		us.LastText = ""
		contextLen := len(us.Context)

		mc.addRequestToPool(us.ID, req)
		defer mc.requestPool.Delete(us.ID)

		chatMessage := newChatMessage(us.Dialog.ID, contextLen, store.RoleUser, text)
		_, err = mc.store.AddNewMessage(req.Ctx, us, chatMessage)
		if err != nil {
			msgEx.sendError(fmt.Errorf("%s: %w", method, err))
			return
		}

		var messagesToAi []ai.Message
		for _, v := range us.Context {
			role := ""
			switch v.Role {
			case store.RoleAssistant:
				role = "assistant"
			case store.RoleSystem:
				role = "system"
			case store.RoleUser:
				role = "user"
			}
			message := ai.Message{Role: role, Content: v.Content}
			messagesToAi = append(messagesToAi, message)
		}

		aiRequest := ai.ChatRequest{
			Model:    "gpt-4o-mini",
			Stream:   true,
			Messages: messagesToAi,
			User:     fmt.Sprint(us.ID),
		}

		answer, err := mc.aiAPI.GetStreamMessages(aiRequest)
		if err != nil {
			msgEx.sendError(fmt.Errorf("%s: %w", method, err))
			return
		}

		sentMsg, err := msgEx.send(newTgMessage(us.ID, "..."))
		if err != nil {
			msgEx.sendError(err)
			return
		}

		var wg sync.WaitGroup
		wg.Add(1)
		var sb strings.Builder
		var totalSb strings.Builder
		requestCanceled := false

		go func() {
			defer wg.Done()
			ticker := time.NewTicker(TgSendingMessageFrequency)
			defer ticker.Stop()
			for {
				if requestCanceled || req.AICtx.Err() != nil {
					requestCanceled = true
					return
				}

				select {
				case <-req.AICtx.Done():
					requestCanceled = true
					return
				case <-ticker.C:
					if sb.Len() > TgMessageMaxLength {
						sentMsg, _ = msgEx.send(newTgEditMessage(us.ID, sentMsg.MessageID, sb.String()+"..."))
						totalSb.WriteString(sb.String())
						sb.Reset()
						sb.WriteString("...")
						sentMsg, _ = msgEx.send(newTgMessage(us.ID, "..."))
						continue
					}
					msgToSend := newTgEditMessage(us.ID, sentMsg.MessageID, sb.String()+"...")
					msgToSend.ReplyMarkup = kbWithOneButton(
						"❌",
						localeText(us.Locale, localization.MTypeBtnCancelRequest),
						fmt.Sprint(callbackTypeCancelRequest))
					sentMsg, _ = msgEx.send(newTgMessage(us.ID, "..."))
				default:
					txt, ok := <-answer
					sb.WriteString(txt)
					if !ok {
						return
					}
				}
			}
		}()

		wg.Wait()

		if requestCanceled {
			sb.WriteString("\n\n----------\n" + localeText(us.Locale, localization.MTypeMsgRequestCanceledByUser))
		}

		msgToSend := newTgEditMessage(us.ID, sentMsg.MessageID, sb.String())
		sentMsg, _ = msgEx.send(msgToSend)

		totalSb.WriteString(sb.String())
		aiChatMessage := &store.ChatMessage{DialogID: us.Dialog.ID, Order: contextLen + 1, Role: store.RoleAssistant, Content: totalSb.String(), Created: time.Now().UTC()}
		_, err = mc.store.AddNewMessage(req.Ctx, us, aiChatMessage)
		if err != nil {
			msgEx.sendError(fmt.Errorf("%s: %w", method, err))
			return
		}

		err = mc.store.UpdateUserUsage(req.Ctx, us, us.User.ChatModelID)
		if err != nil {
			msgEx.sendError(fmt.Errorf("%s: %w", method, err))
			return
		}
	}()

	return msgEx
}

func CheckLastMessageTime(mc *MainController, us *store.UserShell, text string, msgEx *MessageManager) (bool, error) {
	if us.LastText == "" && len(us.Context) > 0 && mc.store.CheckUserLastActivity(us) {
		if us.User.SkipNewDialogMessage {
			err := mc.store.ResetActiveDialog(context.TODO(), us)
			if err != nil {
				return false, err
			}
			return false, nil
		}

		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("❌ %s", localeText(us.Locale, localization.MTypeBtnNo)),
					fmt.Sprint(callbackTypeHandleLastMessage, ";", us.ID, ";", 0)),
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("✅ %s", localeText(us.Locale, localization.MTypeBtnYes)),
					fmt.Sprint(callbackTypeHandleLastMessage, ";", us.ID, ";", 1))))

		msg := newTgMessage(us.ID, localeText(us.Locale, localization.MTypeAnswerCreateNewDialog))
		msg.ReplyMarkup = kb

		sentMsg, _ := msgEx.send(msg)

		us.LastText = text
		us.InfoMessageID = sentMsg.MessageID
		return true, nil
	}
	return false, nil
}

func dialogTitle(str string) string {
	r := []rune(str)
	if len(r) > 30 {
		return string(r[:30])
	}
	return str
}
