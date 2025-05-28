package localization

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Lang string

const (
	LangEN      Lang   = "en"
	LangRU      Lang   = "ru"
	LocalesPath string = "./internal/localization/locales/"
)

type MessageType string

const (
	MTypeMsgStart                     MessageType = "msg_start"
	MTypeMsgCommonError               MessageType = "msg_common_error"
	MTypeMsgNewDialogCreated          MessageType = "msg_new_dialog_created"
	MTypeMsgDialogSelected            MessageType = "msg_dialog_selected"
	MTypeMsgDialogDeleted             MessageType = "msg_dialog_deleted"
	MTypeMsgAllDialogsDeleted         MessageType = "msg_all_dialogs_deleted"
	MTypeMsgYou                       MessageType = "msg_you"
	MTypeMsgAssistant                 MessageType = "msg_assistant"
	MTypeMsgWaitPreviousRequest       MessageType = "msg_wait_previous_request"
	MTypeMsgRequestCanceledByUser     MessageType = "msg_request_canceled_by_user"
	MTypeMsgYourDialogs               MessageType = "msg_your_dialogs"
	MTypeMsgYourHaveNoDialogs         MessageType = "msg_you_have_no_dialogs"
	MTypeMsgProfile                   MessageType = "msg_profile"
	MTypeMsgLimitReached              MessageType = "msg_limit_reached"
	MTypeMsgMaintenance               MessageType = "msg_maintenance"
	MTypeBtnViewAllMessages           MessageType = "btn_view_all_messages"
	MTypeBtnDeleteDialog              MessageType = "btn_delete_dialog"
	MTypeBtnCancel                    MessageType = "btn_cancel"
	MTypeBtnDelete                    MessageType = "btn_delete"
	MTypeBtnYes                       MessageType = "btn_yes"
	MTypeBtnNo                        MessageType = "btn_no"
	MTypeBtnCancelRequest             MessageType = "btn_cancel_request"
	MTypeBtnCancelPreviousRequest     MessageType = "btn_cancel_previous_request"
	MTypeBtnDeleteAllDialogs          MessageType = "btn_delete_all_dialogs"
	MTypeBtnToggleNewDialog           MessageType = "btn_toggle_new_dialog"
	MTypeAnswerDeleteDialog           MessageType = "answer_delete_dialog"
	MTypeAnswerDeleteAllDialogs       MessageType = "answer_delete_all_dialogs"
	MTypeAnswerCreateNewDialog        MessageType = "answer_create_new_dialog"
	MTypeNotifyNoMoreDialogs          MessageType = "notify_no_more_dialogs"
	MTypeNotifyRequestCanceled        MessageType = "notify_request_canceled"
	MTypeNotifyRequestAlreadyCanceled MessageType = "notify_request_already_canceled"
)

var messages = map[Lang]map[MessageType]string{}

func MustLoadMessages(base Lang) {
	langs := []Lang{LangEN, LangRU}

	for _, lang := range langs {
		path := fmt.Sprintf("%s%s.yaml", LocalesPath, lang)
		data, err := os.ReadFile(path)
		if err != nil {
			panic(fmt.Sprintf("cannot read file %s: %v", path, err))
		}

		var m map[MessageType]string
		if err := yaml.Unmarshal(data, &m); err != nil {
			panic(fmt.Sprintf("cannot parse YAML %s: %v", path, err))
		}

		messages[lang] = m
	}

	for _, key := range allMessageTypes() {
		if _, ok := messages[base][MessageType(key)]; !ok {
			panic(fmt.Sprintf("base locale '%s' missing message key: '%s'", base, key))
		}
	}

	// Проверка наличия ключей во всех локалях
	baseMessages, ok := messages[base]
	if !ok {
		panic(fmt.Sprintf("base locale '%s' not loaded", base))
	}

	for lang, m := range messages {
		if lang == base {
			continue
		}

		for key := range baseMessages {
			if _, ok := m[key]; !ok {
				panic(fmt.Sprintf("missing key '%s' in locale '%s'", key, lang))
			}
		}
	}
}

func Message(lang Lang, mType MessageType, args ...any) string {
	msg := messages[lang][mType]

	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

func allMessageTypes() []MessageType {
	return []MessageType{
		MTypeMsgStart,
		MTypeMsgCommonError,
		MTypeMsgNewDialogCreated,
		MTypeMsgDialogSelected,
		MTypeMsgDialogDeleted,
		MTypeMsgAllDialogsDeleted,
		MTypeMsgYou,
		MTypeMsgAssistant,
		MTypeMsgWaitPreviousRequest,
		MTypeMsgRequestCanceledByUser,
		MTypeMsgYourDialogs,
		MTypeMsgYourHaveNoDialogs,
		MTypeMsgProfile,
		MTypeMsgLimitReached,
		MTypeMsgMaintenance,
		MTypeBtnViewAllMessages,
		MTypeBtnDeleteDialog,
		MTypeBtnCancel,
		MTypeBtnDelete,
		MTypeBtnYes,
		MTypeBtnNo,
		MTypeBtnCancelRequest,
		MTypeBtnCancelPreviousRequest,
		MTypeBtnDeleteAllDialogs,
		MTypeAnswerDeleteDialog,
		MTypeAnswerDeleteAllDialogs,
		MTypeAnswerCreateNewDialog,
		MTypeNotifyNoMoreDialogs,
		MTypeNotifyRequestCanceled,
		MTypeNotifyRequestAlreadyCanceled,
		MTypeBtnToggleNewDialog,
	}
}
