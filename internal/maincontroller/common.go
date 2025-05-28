package maincontroller

import (
	"fmt"
	"log"
	"strings"
	"tgbot/internal/localization"
	"tgbot/internal/store"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func localeText(locale string, mType localization.MessageType, args ...any) string {
	return localization.Message(localization.Lang(locale), mType, args...)
}

func newTgMessage(userID int64, text string) tgbotapi.MessageConfig {
	res := tgbotapi.NewMessage(userID, prepareTxtToTgMarkdown(text))
	res.ParseMode = tgbotapi.ModeMarkdownV2
	return res
}

func newTgEditMessage(userID int64, messageID int, text string) tgbotapi.EditMessageTextConfig {
	res := tgbotapi.NewEditMessageText(userID, messageID, prepareTxtToTgMarkdown(text))
	res.ParseMode = tgbotapi.ModeMarkdownV2
	return res
}

type MdType string

const (
	mdTypeNone          MdType = ""
	mdTypeBold          MdType = "::bold::"
	mdTypeItalic        MdType = "::italic::"
	mdTypeStrikethrough MdType = "::strikethrough::"
	mdTypeMono          MdType = "::mono::"
	mdTypeMonorows      MdType = "::monorows::"
)

func prepareTxtToTgMarkdown(text string) string {
	var result strings.Builder
	runes := []rune(text)
	l := len(runes)
	prevEscape := mdTypeNone

	for i := 0; i < l; {
		//fmt.Println("i", i, "l", l, "runes[i]", string(runes[i]), "prevEscape", prevEscape)
		if runes[i] == '\n' && prevEscape != mdTypeNone && prevEscape != mdTypeMonorows {
			var repl string
			if prevEscape == mdTypeBold {
				repl = "**"
			}
			if prevEscape == mdTypeItalic {
				repl = "*"
			}
			if prevEscape == mdTypeStrikethrough {
				repl = "~~"
			}
			if prevEscape == mdTypeMono {
				repl = "`"
			}
			tmp := result.String()
			result.Reset()
			result.WriteString(ReplaceLast(tmp, string(prevEscape), repl))
			prevEscape = mdTypeNone
			result.WriteRune(runes[i])
			i++
			continue
		}

		// monorows
		if i+2 < l && runes[i] == '`' && runes[i+1] == '`' && runes[i+2] == '`' {
			if prevEscape == mdTypeNone { // open
				prevEscape = mdTypeMonorows
				result.WriteString(string(mdTypeMonorows))
			} else if prevEscape == mdTypeMonorows { // close
				prevEscape = ""
				result.WriteString(string(mdTypeMonorows))
			} else { // if inactive
				result.WriteRune(runes[i])
			}
			i += 3
			continue
		}

		// bold
		if i+1 < l && runes[i] == '*' && runes[i+1] == '*' {
			if prevEscape == mdTypeNone { // open
				prevEscape = mdTypeBold
				result.WriteString(string(mdTypeBold))
			} else if prevEscape == mdTypeBold { // close
				prevEscape = ""
				result.WriteString(string(mdTypeBold))
			} else { // if inactive
				result.WriteRune(runes[i])
				result.WriteRune(runes[i+1])
			}
			i += 2
			continue
		}

		// italic
		if runes[i] == '*' {
			if prevEscape == mdTypeNone { // open
				prevEscape = mdTypeItalic
				result.WriteString(string(mdTypeItalic))
			} else if prevEscape == mdTypeItalic { // close
				prevEscape = ""
				result.WriteString(string(mdTypeItalic))
			} else { // if inactive
				result.WriteRune(runes[i])
			}
			i += 1
			continue
		}

		// strikethrough
		if i+1 < l && runes[i] == '~' && runes[i+1] == '~' {
			if prevEscape == mdTypeNone { // open
				prevEscape = mdTypeStrikethrough
				result.WriteString(string(mdTypeStrikethrough))
			} else if prevEscape == mdTypeStrikethrough { // close
				prevEscape = ""
				result.WriteString(string(mdTypeStrikethrough))
			} else { // if inactive
				result.WriteRune(runes[i])
			}
			i += 2
			continue
		}

		// mono
		if runes[i] == '`' {
			if prevEscape == mdTypeNone { // open
				prevEscape = mdTypeMono
				result.WriteString(string(mdTypeMono))
			} else if prevEscape == mdTypeMono { // close
				prevEscape = ""
				result.WriteString(string(mdTypeMono))
			} else { // if inactive
				result.WriteRune(runes[i])
			}
			i += 1
			continue
		}

		result.WriteRune(runes[i])
		i++
	}

	if prevEscape != "" {
		if prevEscape == mdTypeMonorows {
			result.WriteString(string(mdTypeMonorows))
		} else {
			var repl string
			if prevEscape == mdTypeBold {
				repl = "**"
			}
			if prevEscape == mdTypeItalic {
				repl = "*"
			}
			if prevEscape == mdTypeStrikethrough {
				repl = "~~"
			}
			if prevEscape == mdTypeMono {
				repl = "`"
			}
			tmp := result.String()
			result.Reset()
			result.WriteString(ReplaceLast(tmp, string(prevEscape), repl))
		}
	}

	resultText := result.String()
	specialChars := []string{
		"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!",
	}
	for _, char := range specialChars {
		resultText = strings.ReplaceAll(resultText, char, "\\"+char)
	}

	//fmt.Println(resultText)

	resultText = strings.ReplaceAll(resultText, string(mdTypeBold), "*")
	resultText = strings.ReplaceAll(resultText, string(mdTypeItalic), "_")
	resultText = strings.ReplaceAll(resultText, string(mdTypeStrikethrough), "~")
	resultText = strings.ReplaceAll(resultText, string(mdTypeMono), "`")
	resultText = strings.ReplaceAll(resultText, string(mdTypeMonorows), "```")

	return resultText
}

func ReplaceLast(s, old, new string) string {
	index := strings.LastIndex(s, old)
	if index == -1 {
		return s
	}
	return s[:index] + new + s[index+len(old):]
}

func kbWithOneButton(emoji, text, data string) *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.InlineKeyboardMarkup{}
	kb.InlineKeyboard = append(kb.InlineKeyboard,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", emoji, text), data)))
	return &kb
}

func preparePrice(cents int64, label string) string {
	units := cents / 100
	return fmt.Sprintf("%d%s", units, label)
}

func prepareAllDialogs(handler *MainController, req *Request, offset int) (string, *tgbotapi.InlineKeyboardMarkup, error) {
	allDialogsLen, dialogs, err := handler.store.UserDialogs(req.Ctx, &store.DialogFilter{UserID: &req.UserShell.ID, Limit: 5, Offset: offset})
	if err != nil {
		log.Print(err)
		return "", nil, err
	}

	var curDialogID int64
	curDialogID = -1
	if req.UserShell.Dialog != nil {
		curDialogID = req.UserShell.Dialog.ID
	}

	text := localeText(req.UserShell.Locale, localization.MTypeMsgYourDialogs, offset+1, min(allDialogsLen, offset+5), allDialogsLen)
	kb := tgbotapi.InlineKeyboardMarkup{}
	for _, v := range dialogs {
		kb.InlineKeyboard = append(kb.InlineKeyboard,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprint(
						dialogEmoji(curDialogID, v.ID), " ", v.Title),
					fmt.Sprint(callbackTypeDialog, ";", req.UserShell.ID, ";", v.ID))))
	}

	if allDialogsLen > 5 {
		var prevBtn tgbotapi.InlineKeyboardButton
		var nextBtn tgbotapi.InlineKeyboardButton

		if offset == 0 {
			prevBtn = tgbotapi.NewInlineKeyboardButtonData("⏹️", fmt.Sprint(callbackTypeNotify, ";", callbackNotifyTypeNoMoreDialogs))
		} else {
			prevBtn = tgbotapi.NewInlineKeyboardButtonData("⬅️", fmt.Sprint(callbackTypeDialogList, ";", req.UserShell.ID, ";", offset-5))
		}

		if offset+5 >= allDialogsLen {
			nextBtn = tgbotapi.NewInlineKeyboardButtonData("⏹️", fmt.Sprint(callbackTypeNotify, ";", callbackNotifyTypeNoMoreDialogs))
		} else {
			nextBtn = tgbotapi.NewInlineKeyboardButtonData("➡️", fmt.Sprint(callbackTypeDialogList, ";", req.UserShell.ID, ";", offset+5))
		}

		kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(prevBtn, nextBtn))
	}

	if len(kb.InlineKeyboard) > 0 {
		kb.InlineKeyboard = append(kb.InlineKeyboard,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("❌ %s", localeText(req.UserShell.Locale, localization.MTypeBtnDeleteAllDialogs)),
					fmt.Sprint(callbackTypeRemoveAllDialogs, ";", req.UserShell.ID))))
	} else {
		text = localeText(req.UserShell.Locale, localization.MTypeMsgYourHaveNoDialogs)
	}
	return text, &kb, nil
}

func dialogEmoji(curDialogID int64, dialogID int64) string {
	if curDialogID != dialogID {
		return "💬"
	}
	return "✅"
}

func newChatMessage(dialogID int64, order int, role store.ChatMessageRole, content string) *store.ChatMessage {
	return &store.ChatMessage{
		DialogID: dialogID,
		Order:    order,
		Role:     role,
		Content:  content,
		Created:  time.Now().UTC(),
	}
}

func fixedSentFrom(update *tgbotapi.Update) *tgbotapi.User {
	switch {
	case update.MyChatMember != nil:
		return &update.MyChatMember.From
	case update.ChatMember != nil:
		return &update.ChatMember.From
	case update.ChatJoinRequest != nil:
		return &update.ChatJoinRequest.From
	default:
		return update.SentFrom()
	}
}

func prepareProfileMessage(mc *MainController, us *store.UserShell) (string, *tgbotapi.InlineKeyboardMarkup, error) {
	tariff, ok := mc.store.TariffByID(us.User.TariffID)
	if !ok {
		return "", nil, fmt.Errorf("handleCommandProfile(): %w", store.ErrIncorrectTariff)
	}

	var userUsage []*store.UserUsage
	us.Usage.Range(func(_, v any) bool {
		if val, ok := v.(*store.UserUsage); ok {
			userUsage = append(userUsage, val)
		}
		return true
	})

	text := localeText(
		us.Locale,
		localization.MTypeMsgProfile,
		us.ID,
		tariff.Tariff.Title,
		mc.store.UserUsageToString(tariff.Limits, userUsage))

	em := "✅"
	if us.User.SkipNewDialogMessage {
		em = "❌"
	}

	kb := kbWithOneButton(em,
		localeText(us.Locale, localization.MTypeBtnToggleNewDialog),
		fmt.Sprint(callbackTypeToggleNewDialog, ";", us.ID))
	return text, kb, nil
}
