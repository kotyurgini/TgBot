package maincontroller

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TgMessage struct {
	msg       tgbotapi.Chattable
	replyChan chan *TgReplyMessage
}

type TgReplyMessage struct {
	msg *tgbotapi.Message
	err error
}

func prepareTgMessage(msg tgbotapi.Chattable) *TgMessage {
	return &TgMessage{
		msg:       msg,
		replyChan: nil,
	}
}

func prepareTgMessageWithReply(msg tgbotapi.Chattable) *TgMessage {
	replyChan := make(chan *TgReplyMessage)
	return &TgMessage{
		msg:       msg,
		replyChan: replyChan,
	}
}

func prepareTgReplyMessage(rMsg *tgbotapi.Message) *TgReplyMessage {
	return &TgReplyMessage{
		msg: rMsg,
		err: nil,
	}
}

func prepareTgReplyMessageWithErr(err error) *TgReplyMessage {
	return &TgReplyMessage{
		msg: nil,
		err: err,
	}
}
