package store

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrDBNoFilterProvided = errors.New("no filter provided")
	ErrDBNoRowsAffected   = errors.New("no rows affected")
	ErrDBQueryError       = errors.New("query error")
	ErrDBRowError         = errors.New("row error")
	ErrDBScanRowError     = errors.New("scan row error")
)

type Driver interface {
	GetDB() *sql.DB
	Close() error

	// AppSettings
	AppSettingGet(ctx context.Context) (*AppSetting, error)
	AppSettingUpdate(ctx context.Context, entity *AppSetting) (*AppSetting, error)

	// Users
	UserCreate(ctx context.Context, entity *User) (*User, error)
	UserList(ctx context.Context, filter *UserFilter) ([]*User, error)
	UserUpdate(ctx context.Context, entity *User) (*User, error)
	UserDelete(ctx context.Context, filter *UserFilter) error

	// Dialogs
	DialogCreate(ctx context.Context, entity *Dialog) (*Dialog, error)
	DialogList(ctx context.Context, filter *DialogFilter) (int, []*Dialog, error)
	DialogUpdate(ctx context.Context, entity *Dialog) (*Dialog, error)
	DialogDelete(ctx context.Context, entity *DialogFilter) error

	// ActiveDialogs
	ActiveDialogUpsert(ctx context.Context, entity *ActiveDialog) (*ActiveDialog, error)
	ActiveDialogList(ctx context.Context, filter *ActiveDialogFilter) ([]*ActiveDialog, error)
	ActiveDialogDelete(ctx context.Context, entity *ActiveDialogFilter) error

	// ChatMessages
	ChatMessageCreate(ctx context.Context, entity *ChatMessage) (*ChatMessage, error)
	ChatMessageList(ctx context.Context, filter *ChatMessageFilter) ([]*ChatMessage, error)
	ChatMessageUpdate(ctx context.Context, entity *ChatMessage) (*ChatMessage, error)
	ChatMessageDelete(ctx context.Context, entity *ChatMessageFilter) error

	// AiModels
	AiModelList(ctx context.Context) ([]*AiModel, error)

	// Tariffs
	TariffCreate(ctx context.Context, entity *Tariff) (*Tariff, error)
	TariffList(ctx context.Context, filter *TariffFilter) ([]*Tariff, error)
	TariffUpdate(ctx context.Context, entity *Tariff) (*Tariff, error)
	TariffDelete(ctx context.Context, filter *TariffFilter) error

	// TariffsLimit
	TariffLimitCreate(ctx context.Context, entity *TariffLimit) (*TariffLimit, error)
	TariffLimitList(ctx context.Context, filter *TariffLimitFilter) ([]*TariffLimit, error)
	TariffLimitUpdate(ctx context.Context, entity *TariffLimit) (*TariffLimit, error)
	TariffLimitDelete(ctx context.Context, filter *TariffLimitFilter) error

	// UsersUsage
	UserUsageCreate(ctx context.Context, entity *UserUsage) (*UserUsage, error)
	UserUsageList(ctx context.Context, filter *UserUsageFilter) ([]*UserUsage, error)
	UserUsageUpdate(ctx context.Context, entity *UserUsage) (*UserUsage, error)
	UserUsageDelete(ctx context.Context, filter *UserUsageFilter) error
}
