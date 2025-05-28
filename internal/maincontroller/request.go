package maincontroller

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"tgbot/internal/store"
	"time"
)

type Request struct {
	Ctx       context.Context
	Cancel    context.CancelFunc `json:"-"`
	AICtx     context.Context
	AICancel  context.CancelFunc `json:"-"`
	Update    *tgbotapi.Update
	StartTime time.Time
	UserShell *store.UserShell
}

func newRequest(update *tgbotapi.Update) *Request {
	ctx, cancel := context.WithCancel(context.Background())
	aiCtx, aiCancel := context.WithCancel(context.Background())
	return &Request{
		Ctx:       ctx,
		Cancel:    cancel,
		Update:    update,
		StartTime: time.Now().UTC(),
		AICtx:     aiCtx,
		AICancel:  aiCancel,
	}
}
