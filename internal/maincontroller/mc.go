package maincontroller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"tgbot/internal/ai"
	"tgbot/internal/localization"
	"tgbot/internal/store"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MainController struct {
	tgBot       *tgbotapi.BotAPI
	Ctx         context.Context
	store       *store.Store
	aiAPI       ai.ChatModel
	log         *slog.Logger
	requestPool sync.Map // [UserId] *Request
	tgAdmin     int64
}

var (
	errSendErrorMessage    = errors.New("error while send error message")
	errMaintenanceModeIsOn = errors.New("maintenance mode is on")
	errUserBlockedBot      = errors.New("user blocked bot")
	errUserBlocked         = errors.New("user blocked")
)

const (
	TgMessageMaxLength        int = 3700 // symbols
	TgSendingMessageFrequency     = 2000 * time.Millisecond
)

func New(ctx context.Context, tgBot *tgbotapi.BotAPI, st *store.Store, aiAPI ai.ChatModel, log *slog.Logger, tgAdmin int64) (*MainController, error) {
	mc := MainController{tgBot: tgBot, Ctx: ctx, store: st, aiAPI: aiAPI, log: log, tgAdmin: tgAdmin}

	// for i := range 25 {
	// 	_, err := st.AddDialog(context.Background(), &store.UserShell{ID: tgAdmin}, &store.Dialog{Title: fmt.Sprintf("Test dialog %d", i), UserID: tgAdmin})
	// 	if err != nil {
	// 		fmt.Printf("%v\n", err)
	// 		break
	// 	}
	// }

	go func() {
		uConf := tgbotapi.NewUpdate(0)
		uConf.Timeout = 60
		tgUpdates := tgBot.GetUpdatesChan(uConf)
		for {
			if mc.Ctx.Err() != nil {
				return
			}
			select {
			case update := <-tgUpdates:
				go mc.handleTgUpdate(&update)
			case <-mc.Ctx.Done():
				return
			}
		}
	}()

	return &mc, nil
}

func (mc *MainController) handleTgUpdate(update *tgbotapi.Update) {
	tgUser := fixedSentFrom(update)
	startTime := time.Now()
	mc.log.Info("Get update",
		slog.Attr{Key: "Update id", Value: slog.IntValue(update.UpdateID)},
		slog.Attr{Key: "User id", Value: slog.Int64Value(tgUser.ID)})

	var err error
	if mc.CheckMaintenance() {
		if mc.itsAdmin(tgUser.ID) {
			msg := newTgMessage(tgUser.ID, "Включен режим обслуживания")
			_, _ = mc.sendMessageToTgBot(nil, msg)
		} else {
			msg := newTgMessage(tgUser.ID, localeText(tgUser.LanguageCode, localization.MTypeMsgMaintenance))
			_, _ = mc.sendMessageToTgBot(nil, msg)
			mc.handledLog(errMaintenanceModeIsOn, update.UpdateID, startTime)
			return
		}
	}

	req := newRequest(update)
	user, err := mc.store.ValidateUser(req.Ctx, tgUser.ID, tgUser.LanguageCode)
	if err != nil {
		mc.handledLog(err, update.UpdateID, startTime)
		return
	}
	req.UserShell = user

	if user.User.Blocked {
		mc.handledLog(errUserBlocked, update.UpdateID, startTime)
		return
	}

	mc.store.SetSelfBlockUser(user, false)

	var msgEx *MessageManager
	switch {
	case req.Update.Message != nil:
		if req.Update.Message.IsCommand() {
			msgEx = mc.handleTgCommand(req)
		} else {
			msgEx = mc.handleTgMessage(req)
		}
	case req.Update.CallbackQuery != nil:
		msgEx = mc.handleTgCallback(req)
	case req.Update.MyChatMember != nil:
		err = mc.handleTgMyChatMember()
	}

	var sentMessage tgbotapi.Message
	var msg tgbotapi.Chattable
	var ok bool
	for {
		if msgEx.canceled() {
			break
		}
		select {
		case msg, ok = <-msgEx.msgChan:
			if !ok {
				break
			}
			sentMessage, err = mc.sendMessageToTgBot(req.UserShell, msg)
			if err != nil {
				msgEx.replyErrorChan <- err
			} else {
				msgEx.replyChan <- &sentMessage
			}
		case err, ok = <-msgEx.errorChan:
			if !ok {
				break
			}
			// errors to ignore for user
			if !errors.Is(err, ErrPermissionDenied) && !errors.Is(err, errUserBlockedBot) && !errors.Is(err, ErrCommandNotFound) {
				err = mc.sendErrorToTgBot("handle tg update()", user, err)
			}
		default:
		}
	}
	mc.handledLog(err, update.UpdateID, startTime)
}

func (mc *MainController) handleTgMyChatMember() error {
	mc.log.Info("handleTgMyChatMember()")
	return nil
}

func (mc *MainController) handledLog(err error, updateID int, startTime time.Time) {
	attrs := []any{
		slog.Attr{Key: "Update id", Value: slog.IntValue(updateID)},
		slog.Attr{Key: "Time", Value: slog.StringValue(time.Since(startTime).String())},
	}

	if err == nil {
		attrs = append(attrs, slog.Attr{Key: "Handle result", Value: slog.BoolValue(true)})
	} else {
		attrs = append(attrs, slog.Attr{Key: "Handle result", Value: slog.BoolValue(false)})
		attrs = append(attrs, slog.Attr{Key: "Error", Value: slog.StringValue(err.Error())})
	}

	mc.log.Info("Update handled", attrs...)
}

func (mc *MainController) CheckMaintenance() bool {
	return mc.store.MaintenanceStatus()
}

func (mc *MainController) sendMessageToTgBot(user *store.UserShell, message tgbotapi.Chattable) (tgbotapi.Message, error) {
	if user != nil && user.User.SelfBlock {
		return tgbotapi.Message{}, fmt.Errorf("sendMessageToTgBot(): %w", errUserBlockedBot)
	}

	if msg, ok := message.(tgbotapi.CallbackConfig); ok {
		_, _ = mc.tgBot.Request(msg)
		return tgbotapi.Message{}, nil
	}
	if msg, ok := message.(tgbotapi.DeleteMessageConfig); ok {
		_, _ = mc.tgBot.Request(msg)
		return tgbotapi.Message{}, nil
	}

	msg, err := mc.tgBot.Send(message)
	if err != nil && strings.Contains(err.Error(), "bot was blocked by the user") {
		mc.store.SetSelfBlockUser(user, true)
		return tgbotapi.Message{}, fmt.Errorf("sendMessageToTgBot1(): %w", errUserBlockedBot)
	}
	if err != nil {
		return tgbotapi.Message{}, fmt.Errorf("sendMessageToTgBot(): %w", err)
	}
	return msg, nil
}

func (mc *MainController) sendErrorToTgBot(method string, userShell *store.UserShell, err error) error {
	if err == nil {
		return nil
	}

	_, sendErr := mc.sendMessageToTgBot(userShell, newTgMessage(userShell.ID, localeText(userShell.Locale, localization.MTypeMsgCommonError)))
	if sendErr != nil {
		return fmt.Errorf("%s: %w: %w", method, errSendErrorMessage, sendErr)
	}
	return fmt.Errorf("%s: %w", method, err)
}

func (mc *MainController) addRequestToPool(userID int64, req *Request) {
	mc.requestPool.Store(userID, req)
}

func (mc *MainController) hasActiveUserRequest(userID int64) (*Request, bool) {
	request, ok := mc.requestPool.Load(userID)
	if ok {
		return request.(*Request), ok
	}
	return nil, ok
}

func (mc *MainController) cancelPreviousRequest(userID int64) {
	if req, ok := mc.requestPool.LoadAndDelete(userID); ok {
		req.(*Request).AICancel()
	}
}

func (mc *MainController) itsAdmin(userID int64) bool {
	return userID == mc.tgAdmin
}
