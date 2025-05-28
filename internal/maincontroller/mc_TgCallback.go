package maincontroller

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"tgbot/internal/localization"
	"tgbot/internal/store"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	errInvalidCallbackData     = errors.New("invalid callback data")
	errInvalidCallbackType     = errors.New("invalid callback type")
	errFailedMatchCallbackType = errors.New("failed to match type")
)

type CallbackType int

const (
	callbackTypeDialog CallbackType = iota
	callbackTypeDialogList
	callbackTypeRemoveDialog
	callbackTypeRemoveAllDialogs
	callbackTypeNotify
	callbackTypeAllMessages
	callbackTypeHandleLastMessage
	callbackTypeCancelRequest
	callbackTypeTariff
	callbackTypeToggleNewDialog
)

type CallbackNotifyType int

const (
	callbackNotifyTypeNoMoreDialogs CallbackNotifyType = iota
	callbackNotifyRequestCanceled
	callbackNotifyRequestAlreadyCancelled
)

func (mc *MainController) handleTgCallback(req *Request) *MessageManager {
	method := "handleTgCallback()"

	msgEx := newMessageExchange()
	go func() {
		defer msgEx.close()

		data := strings.Split(req.Update.CallbackQuery.Data, ";")

		if len(data) < 1 {
			msgEx.sendError(fmt.Errorf("%s: %w", method, errInvalidCallbackData))
			return
		}

		cbType, err := strconv.Atoi(data[0])
		if err != nil {
			msgEx.sendError(fmt.Errorf("%s: %w", method, errInvalidCallbackType))
			return
		}

		switch CallbackType(cbType) {
		case callbackTypeDialog:
			handleCallbackDialog(mc, req, data, msgEx)
		case callbackTypeDialogList:
			handleCallbackDialogList(mc, req, data, msgEx)
		case callbackTypeRemoveDialog:
			handleCallbackRemoveDialog(mc, req, data, msgEx)
		case callbackTypeRemoveAllDialogs:
			handleCallbackRemoveAllDialogs(mc, req, data, msgEx)
		case callbackTypeNotify:
			handleCallbackNotify(req, data, msgEx)
		case callbackTypeAllMessages:
			handleCallbackAllMessages(req, msgEx)
		case callbackTypeHandleLastMessage:
			handleCallbackHandleLastMessage(mc, req, data, msgEx)
		case callbackTypeCancelRequest:
			handleCallbackCancelRequest(mc, req, msgEx)
		case callbackTypeTariff:
			handleCallbackTariff(mc, req, data, msgEx)
		case callbackTypeToggleNewDialog:
			handleCallbackTypeToggleNewDialog(mc, req, data, msgEx)
		default:
			msgEx.sendError(fmt.Errorf("%s: %w", method, errFailedMatchCallbackType))
			return
		}
	}()
	return msgEx
}

func notifyMessage(t CallbackNotifyType, locale string) string {
	switch t {
	case callbackNotifyTypeNoMoreDialogs:
		return localeText(locale, localization.MTypeNotifyNoMoreDialogs)
	case callbackNotifyRequestAlreadyCancelled:
		return localeText(locale, localization.MTypeNotifyRequestAlreadyCanceled)
	case callbackNotifyRequestCanceled:
		return localeText(locale, localization.MTypeNotifyRequestCanceled)
	}
	return "---"
}

func checkDataLen(data []string, minLen int, method string) error {
	if len(data) < minLen {
		return fmt.Errorf("%s: %w", method, errInvalidCallbackData)
	}
	return nil
}

func handleCallbackDialog(mc *MainController, req *Request, data []string, msgEx *MessageManager) {
	method := "handleCallbackDialog()"

	if err := checkDataLen(data, 2, method); err != nil {
		msgEx.sendError(err)
		return
	}

	dialogID, err := strconv.ParseInt(data[2], 10, 64)
	if err != nil {
		msgEx.sendError(fmt.Errorf("%s: %w", method, err))
		return
	}

	dialogInfo, err := mc.store.DialogInfo(req.Ctx, dialogID)
	if err != nil {
		msgEx.sendError(fmt.Errorf("%s: %w", method, err))
		return
	}

	err = mc.store.UpdateActiveDialog(req.Ctx, req.UserShell, dialogInfo)
	if err != nil {
		msgEx.sendError(fmt.Errorf("%s: %w", method, err))
		return
	}

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("✉️ %s", localeText(req.UserShell.Locale, localization.MTypeBtnViewAllMessages)),
				fmt.Sprint(callbackTypeAllMessages, ";", req.UserShell.ID, ";", dialogInfo.Dialog.ID))),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("❌ %s", localeText(req.UserShell.Locale, localization.MTypeBtnDeleteDialog)),
				fmt.Sprint(callbackTypeRemoveDialog, ";", req.UserShell.ID, ";", dialogInfo.Dialog.ID))))

	msg := newTgEditMessage(req.UserShell.ID, req.Update.CallbackQuery.Message.MessageID, localeText(req.UserShell.Locale, localization.MTypeMsgDialogSelected))
	msg.ReplyMarkup = &kb

	_, _ = msgEx.send(msg)
}

func handleCallbackDialogList(mc *MainController, req *Request, data []string, msgEx *MessageManager) {
	method := "handleCallbackDialogList()"

	if err := checkDataLen(data, 3, method); err != nil {
		msgEx.sendError(err)
		return
	}

	offset, err := strconv.Atoi(data[2])
	if err != nil {
		msgEx.sendError(fmt.Errorf("%s: %w", method, err))
		return
	}

	text, dialogsBtns, err := prepareAllDialogs(mc, req, offset)
	if err != nil {
		msgEx.sendError(fmt.Errorf("%s: %w", method, err))
		return
	}

	msg := newTgEditMessage(req.UserShell.ID, req.Update.CallbackQuery.Message.MessageID, text)
	msg.ReplyMarkup = dialogsBtns

	_, _ = msgEx.send(msg)
}

func handleCallbackRemoveDialog(mc *MainController, req *Request, data []string, msgEx *MessageManager) {
	method := "handleCallbackRemoveDialog()"

	if err := checkDataLen(data, 3, method); err != nil {
		msgEx.sendError(err)
		return
	}

	dialogID, err := strconv.ParseInt(data[2], 10, 64)
	if err != nil {
		msgEx.sendError(fmt.Errorf("%s: %w", method, err))
		return
	}

	var msg tgbotapi.EditMessageTextConfig
	if len(data) < 4 {
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("❌ %s", localeText(req.UserShell.Locale, localization.MTypeBtnCancel)),
					fmt.Sprint(callbackTypeDialog, ";", req.UserShell.ID, ";", dialogID)),
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("✅ %s", localeText(req.UserShell.Locale, localization.MTypeBtnDelete)),
					fmt.Sprint(callbackTypeRemoveDialog, ";", req.UserShell.ID, ";", dialogID, ";", 1)),
			),
		)
		msg = newTgEditMessage(req.UserShell.ID, req.Update.CallbackQuery.Message.MessageID, localeText(req.UserShell.Locale, localization.MTypeAnswerDeleteDialog))
		msg.ReplyMarkup = &kb
	} else {
		err = mc.store.DeleteDialog(req.Ctx, req.UserShell, &store.DialogFilter{ID: &dialogID})
		if err != nil {
			msgEx.sendError(fmt.Errorf("%s: %w", method, err))
			return
		}
		msg = newTgEditMessage(req.UserShell.ID, req.Update.CallbackQuery.Message.MessageID, localeText(req.UserShell.Locale, localization.MTypeMsgDialogDeleted))
	}

	_, _ = msgEx.send(msg)
}

func handleCallbackRemoveAllDialogs(mc *MainController, req *Request, data []string, msgEx *MessageManager) {
	method := "handleCallbackRemoveAllDialogs()"

	if err := checkDataLen(data, 2, method); err != nil {
		msgEx.sendError(err)
		return
	}

	var msg tgbotapi.EditMessageTextConfig
	if len(data) < 3 {
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("❌ %s", localeText(req.UserShell.Locale, localization.MTypeBtnCancel)),
					fmt.Sprint(callbackTypeDialogList, ";", req.UserShell.ID, ";", 0)),
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("✅ %s", localeText(req.UserShell.Locale, localization.MTypeBtnDelete)),
					fmt.Sprint(callbackTypeRemoveAllDialogs, ";", req.UserShell.ID, ";", 1)),
			),
		)
		msg = newTgEditMessage(req.UserShell.ID, req.Update.CallbackQuery.Message.MessageID, localeText(req.UserShell.Locale, localization.MTypeAnswerDeleteAllDialogs))
		msg.ReplyMarkup = &kb
	} else {
		err := mc.store.DeleteDialog(req.Ctx, req.UserShell, &store.DialogFilter{UserID: &req.UserShell.ID})
		if err != nil {
			msgEx.sendError(fmt.Errorf("%s: %w", method, err))
			return
		}
		msg = newTgEditMessage(req.UserShell.ID, req.Update.CallbackQuery.Message.MessageID, localeText(req.UserShell.Locale, localization.MTypeMsgAllDialogsDeleted))
	}

	_, _ = msgEx.send(msg)
}

func handleCallbackNotify(req *Request, data []string, msgEx *MessageManager) {
	method := "handleCallbackNotify()"

	if err := checkDataLen(data, 2, method); err != nil {
		msgEx.sendError(err)
		return
	}

	notifyType, err := strconv.Atoi(data[1])
	if err != nil {
		msgEx.sendError(fmt.Errorf("%s: %w", method, err))
		return
	}
	msg := tgbotapi.NewCallbackWithAlert(req.Update.CallbackQuery.ID, notifyMessage(CallbackNotifyType(notifyType), req.UserShell.Locale))
	msg.ShowAlert = false

	_, _ = msgEx.send(msg)
}

// View all messages in the current user dialog
func handleCallbackAllMessages(req *Request, msgEx *MessageManager) {
	var sb strings.Builder
	for _, v := range req.UserShell.Context {
		prefix := fmt.Sprintf("🧑‍💻 %s: ", localeText(req.UserShell.Locale, localization.MTypeMsgYou))
		if v.Role != store.RoleUser {
			prefix = fmt.Sprintf("🤖 %s: ", localeText(req.UserShell.Locale, localization.MTypeMsgAssistant))
		}
		sb.WriteString(prefix)
		sb.WriteString(v.Content)
		sb.WriteString("\n")
		if sb.Len() > TgMessageMaxLength {
			_, _ = msgEx.send(newTgMessage(req.UserShell.ID, sb.String()))
			sb.Reset()
		}
	}

	if sb.Len() > 0 {
		_, _ = msgEx.send(newTgMessage(req.UserShell.ID, sb.String()))
	}
}

func handleCallbackHandleLastMessage(mc *MainController, req *Request, data []string, msgEx *MessageManager) {
	method := "handleCallbackHandleLastMessage()"
	if err := checkDataLen(data, 3, method); err != nil {
		msgEx.sendError(err)
		return
	}

	_, _ = msgEx.send(tgbotapi.NewDeleteMessage(req.UserShell.ID, req.UserShell.InfoMessageID))

	if data[2] == "1" {
		err := mc.store.ResetActiveDialog(mc.Ctx, req.UserShell)
		if err != nil {
			msgEx.sendError(fmt.Errorf("%s: %w", method, err))
			return
		}
	}

	// TODO
	//err := mc.handleTgMessage(req, ch)
	//if err != nil {
	//	return nil, fmt.Errorf("%s: %w", method, err)
	//}
}

func handleCallbackCancelRequest(mc *MainController, req *Request, msgEx *MessageManager) {
	_, ok := mc.hasActiveUserRequest(req.UserShell.ID)
	if !ok {
		msg := tgbotapi.NewCallbackWithAlert(req.Update.CallbackQuery.ID, notifyMessage(callbackNotifyRequestAlreadyCancelled, req.UserShell.Locale))
		msg.ShowAlert = false
		_, _ = msgEx.send(msg)
	}

	mc.cancelPreviousRequest(req.UserShell.ID)

	msg := tgbotapi.NewCallbackWithAlert(req.Update.CallbackQuery.ID, notifyMessage(callbackNotifyRequestCanceled, req.UserShell.Locale))
	msg.ShowAlert = false

	_, _ = msgEx.send(msg)
}

func handleCallbackTariff(mc *MainController, req *Request, data []string, msgEx *MessageManager) {
	method := "handleCallbackCancelRequest()"
	if err := checkDataLen(data, 2, method); err != nil {
		msgEx.sendError(err)
		return
	}

	tariffID, err := strconv.ParseInt(data[1], 10, 64)
	if err != nil {
		msgEx.sendError(fmt.Errorf("%s: %w", method, err))
		return
	}

	tariff, ok := mc.store.TariffByID(int32(tariffID))
	if !ok {
		msgEx.sendError(fmt.Errorf("%s: %w", method, errors.New("incorrect tariff id")))
		return
	}

	text := fmt.Sprintf("Тариф: %s\n\nДоступные модели:\n", tariff.Tariff.Title)
	for _, limit := range tariff.Limits {
		model, ok := mc.store.AIModelByID(limit.AIModelID)
		if !ok {
			msgEx.sendError(fmt.Errorf("%s: %w", method, errors.New("incorrect ai model id")))
			return
		}
		text += fmt.Sprintf(" - %s (%d)\n", model.Title, limit.Count)
	}

	text += fmt.Sprintf("\nЦены: %v/%v", preparePrice(tariff.Tariff.RubPrice, "р."), preparePrice(tariff.Tariff.UsdPrice, "$"))

	msg := newTgMessage(req.UserShell.ID, text)
	_, _ = msgEx.send(msg)
}

func handleCallbackTypeToggleNewDialog(mc *MainController, req *Request, data []string, msgEx *MessageManager) {
	method := "handleCallbackTypeToggleNewDialog()"
	if err := checkDataLen(data, 2, method); err != nil {
		msgEx.sendError(err)
		return
	}

	err := mc.store.ToggleUserSkipDialog(req.Ctx, req.UserShell)
	if err != nil {
		msgEx.sendError(err)
		return
	}

	text, kb, err := prepareProfileMessage(mc, req.UserShell)
	if err != nil {
		msgEx.sendError(fmt.Errorf("handleCommandProfile(): %w", err))
		return
	}

	msg := newTgEditMessage(req.UserShell.ID, req.Update.CallbackQuery.Message.MessageID, text)
	msg.ReplyMarkup = kb

	_, _ = msgEx.send(msg)
}
