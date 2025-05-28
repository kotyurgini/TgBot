package maincontroller

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"strings"
	"tgbot/internal/localization"
)

type TgCommand string

const (
	CmdStart          TgCommand = "start"
	CmdDialogs        TgCommand = "dialogs"
	CmdNew            TgCommand = "new"
	CmdTariffs        TgCommand = "tariffs"
	CmdProfile        TgCommand = "profile"
	CmdSetMaintenance TgCommand = "setMaintenance"
	CmdBlockUser      TgCommand = "blockUser"
	CmdUnblockUser    TgCommand = "unblockUser"
)

var adminCommands = map[TgCommand]any{
	CmdTariffs:        nil,
	CmdSetMaintenance: nil,
	CmdBlockUser:      nil,
	CmdUnblockUser:    nil,
}

var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrCommandNotFound  = errors.New("command not found")
)

func (mc *MainController) handleTgCommand(req *Request) *MessageManager {
	cmd := TgCommand(req.Update.Message.Command())

	msgEx := newMessageExchange()

	go func() {
		defer msgEx.close()

		if !checkCommandPermissions(mc.tgAdmin, req.UserShell.ID, cmd) {
			msgEx.sendError(fmt.Errorf("handleTgCommand(): %w", ErrPermissionDenied))
			return
		}

		switch cmd {
		case CmdStart:
			handleCommandStart(msgEx, req)
		case CmdDialogs:
			handleCommandDialogs(mc, msgEx, req)
		case CmdNew:
			handleCommandNew(mc, msgEx, req)
		case CmdTariffs:
			handleCommandTariffs(mc, msgEx, req)
		case CmdProfile:
			handleCommandProfile(mc, msgEx, req)
		case CmdSetMaintenance:
			handleCommandSetMaintenance(mc, msgEx, req)
		case CmdBlockUser:
			handleCommandBlockUser(mc, msgEx, req)
		case CmdUnblockUser:
			handleCommandUnblockUser(mc, msgEx, req)
		default:
			msgEx.sendError(fmt.Errorf("handleTgCommand(): %w: '%s'", ErrCommandNotFound, cmd))
		}
	}()

	return msgEx
}

func checkCommandPermissions(adminID, userID int64, cmd TgCommand) bool {
	if _, ok := adminCommands[cmd]; !ok {
		return true
	}
	if adminID != userID {
		return false
	}
	return true
}

func handleCommandStart(msgEx *MessageManager, req *Request) {
	msg := newTgMessage(req.UserShell.ID, localeText(req.UserShell.Locale, localization.MTypeMsgStart))
	_, _ = msgEx.send(msg)
}

func handleCommandDialogs(mc *MainController, msgEx *MessageManager, req *Request) {
	text, dialogsButtons, err := prepareAllDialogs(mc, req, 0)
	if err != nil {
		msgEx.sendError(fmt.Errorf("handleCommandDialogs(): %w", err))
		return
	}

	msg := newTgMessage(req.UserShell.ID, text)
	if len(dialogsButtons.InlineKeyboard) > 0 {
		msg.ReplyMarkup = dialogsButtons
	}
	_, _ = msgEx.send(msg)
}

func handleCommandNew(mc *MainController, msgEx *MessageManager, req *Request) {
	if err := mc.store.ResetActiveDialog(req.Ctx, req.UserShell); err != nil {
		msgEx.sendError(fmt.Errorf("handleCommandNew(): %w", err))
	}

	_, _ = msgEx.send(newTgMessage(req.UserShell.ID, localeText(req.UserShell.Locale, localization.MTypeMsgNewDialogCreated)))
}

func handleCommandTariffs(mc *MainController, msgEx *MessageManager, req *Request) {
	kb := tgbotapi.InlineKeyboardMarkup{}
	tariffs := mc.store.Tariffs()
	for _, tariff := range tariffs {
		kb.InlineKeyboard = append(kb.InlineKeyboard,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprint(tariff.Tariff.Title),
					fmt.Sprint(callbackTypeTariff, ";", tariff.Tariff.ID))))
	}
	msg := newTgMessage(req.UserShell.ID, "Тарифы")
	msg.ReplyMarkup = kb
	_, _ = msgEx.send(msg)
}

func handleCommandProfile(mc *MainController, msgEx *MessageManager, req *Request) {
	text, kb, err := prepareProfileMessage(mc, req.UserShell)
	if err != nil {
		msgEx.sendError(fmt.Errorf("handleCommandProfile(): %w", err))
	}

	msg := newTgMessage(req.UserShell.ID, text)
	msg.ReplyMarkup = kb
	_, _ = msgEx.send(msg)
}

func handleCommandSetMaintenance(mc *MainController, msgEx *MessageManager, req *Request) {
	err := mc.store.SetMaintenance(req.Ctx, !mc.store.MaintenanceStatus())
	if err != nil {
		msgEx.sendError(fmt.Errorf("setMaintenance(): %w", err))
	}

	msg := newTgMessage(req.UserShell.ID, fmt.Sprintf("Maintenance mode: %v", mc.store.MaintenanceStatus()))
	_, _ = msgEx.send(msg)
}

func handleCommandBlockUser(mc *MainController, msgEx *MessageManager, req *Request) {
	method := "handleCommandBlockUser()"
	argsStr := req.Update.Message.CommandArguments()
	args := strings.SplitN(argsStr, " ", 2)
	if len(args) != 2 {
		msgEx.sendError(fmt.Errorf("%s: %w", method, errors.New("invalid arguments")))
	}

	userID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		msgEx.sendError(fmt.Errorf("%s: %w", method, err))
	}

	us, err := mc.store.GetUserShellByID(req.Ctx, userID)
	if err != nil {
		msgEx.sendError(fmt.Errorf("%s: %w", method, err))
	}

	err = mc.store.BlockUser(req.Ctx, us, args[1], mc.tgAdmin)
	if err != nil {
		msgEx.sendError(fmt.Errorf("%s: %w", method, err))
	}

	msg := newTgMessage(req.UserShell.ID, fmt.Sprintf("User %d blocked: %s", userID, args[1]))
	_, _ = msgEx.send(msg)
}

func handleCommandUnblockUser(mc *MainController, msgEx *MessageManager, req *Request) {
	method := "handleCommandUnblockUser()"
	argsStr := req.Update.Message.CommandArguments()
	userID, err := strconv.ParseInt(argsStr, 10, 64)
	if err != nil {
		msgEx.sendError(fmt.Errorf("%s: %w", method, err))
	}

	us, err := mc.store.GetUserShellByID(req.Ctx, userID)
	if err != nil {
		msgEx.sendError(fmt.Errorf("%s: %w", method, err))
	}

	err = mc.store.UnblockUser(req.Ctx, us)
	if err != nil {
		msgEx.sendError(fmt.Errorf("%s: %w", method, err))
	}

	msg := newTgMessage(req.UserShell.ID, fmt.Sprintf("User %d unblocked", userID))
	_, _ = msgEx.send(msg)
}
