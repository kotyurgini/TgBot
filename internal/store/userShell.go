package store

import (
	"sync"
	"time"
)

type UserShell struct {
	User           *User
	ID             int64
	Dialog         *Dialog
	Context        []*ChatMessage
	LastText       string
	InfoMessageID  int
	Locale         string
	LastLimitReset time.Time
	Usage          sync.Map // [int32 modelId]*UserUsage
}
