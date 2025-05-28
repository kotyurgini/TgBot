package maincontroller

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageManager struct {
	msgChan        chan tgbotapi.Chattable
	errorChan      chan error
	replyChan      chan *tgbotapi.Message
	replyErrorChan chan error
	ctx            context.Context
	cancel         context.CancelFunc
}

func newMessageExchange() *MessageManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &MessageManager{
		msgChan:        make(chan tgbotapi.Chattable),
		errorChan:      make(chan error),
		replyChan:      make(chan *tgbotapi.Message),
		replyErrorChan: make(chan error),
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (msgEx *MessageManager) send(msg tgbotapi.Chattable) (*tgbotapi.Message, error) {
	msgEx.msgChan <- msg
	select {
	case rmsg := <-msgEx.replyChan:
		return rmsg, nil
	case err := <-msgEx.replyErrorChan:
		return nil, err
	}
}

func (msgEx *MessageManager) sendError(err error) {
	msgEx.errorChan <- err
}

func (msgEx *MessageManager) close() {
	msgEx.cancel()
	close(msgEx.msgChan)
	close(msgEx.errorChan)
	close(msgEx.replyChan)
	close(msgEx.replyErrorChan)
	msgEx.msgChan = nil
	msgEx.errorChan = nil
	msgEx.replyChan = nil
	msgEx.replyErrorChan = nil
}

func (msgEx *MessageManager) canceled() bool {
	return msgEx.ctx.Err() != nil
}
