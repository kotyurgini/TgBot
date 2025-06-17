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
