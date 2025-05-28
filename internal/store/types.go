package store

import "time"

type AppSetting struct {
	ID          int32
	Maintenance bool
}

type User struct {
	ID                   int64
	ChatModelID          int32
	ImageModelID         int32
	TariffID             int32
	LastLimitReset       time.Time
	SelfBlock            bool
	Blocked              bool
	BlockReason          string
	SkipNewDialogMessage bool
}

type UserFilter struct {
	ID                   *int64
	ChatModelID          *int32
	ImageModelID         *int32
	TariffID             *int32
	LastLimitReset       *time.Time
	SelfBlock            *bool
	Blocked              *bool
	BlockReason          *string
	SkipNewDialogMessage *bool
}

type Dialog struct {
	ID      int64
	UserID  int64
	Title   string
	Created time.Time
}

type DialogFilter struct {
	ID      *int64
	UserID  *int64
	Title   *string
	Created *time.Time
	Limit   int
	Offset  int
}

type DialogInfo struct {
	Dialog  *Dialog
	Context []*ChatMessage
}

type ChatMessageRole int

const (
	RoleSystem ChatMessageRole = iota
	RoleUser
	RoleAssistant
)

type ChatMessage struct {
	ID       int64
	DialogID int64
	Order    int
	Role     ChatMessageRole
	Content  string
	Created  time.Time
}

type ChatMessageFilter struct {
	ID       *int64
	DialogID *int64
	Order    *int
	Role     *ChatMessageRole
	Content  *string
	Created  *time.Time
}

type ActiveDialog struct {
	UserID   int64
	DialogID int64
}

type ActiveDialogFilter struct {
	UserID   *int64
	DialogID *int64
}

type AiModelType int

const (
	TypeChat AiModelType = iota
	TypeGenerateImage
)

type AiModel struct {
	ID        int32
	Title     string
	APIName   string
	ModelType AiModelType
}

type Tariff struct {
	ID        int32
	Title     string
	RubPrice  int64
	UsdPrice  int64
	Available bool
}

type TariffFilter struct {
	ID        *int32
	Title     *string
	RubPrice  *int64
	UsdPrice  *int64
	Available *bool
}

type TariffLimit struct {
	ID        int32
	TariffID  int32
	AIModelID int32
	Count     int32
}

type TariffLimitFilter struct {
	ID        *int32
	TariffID  *int32
	AIModelID *int32
	Count     *int32
}

type UserUsage struct {
	UserID       int64
	AIModelID    int32
	Count        int32
	LastActivity time.Time
}

type UserUsageFilter struct {
	UserID       *int64
	AIModelID    *int32
	Count        *int32
	LastActivity *time.Time
}
