package handler

import (
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gtoxlili/quantifiable-swap/common/config"
	"github.com/gtoxlili/quantifiable-swap/common/smap"
	"github.com/gtoxlili/quantifiable-swap/job"
	"github.com/rs/zerolog"
)

// SessionState holds the user's current interaction state in the bot.
type SessionState struct {
	CurrentAction string
	TempJob       *config.Job
	Step          int
}

// BotHandler orchestrates how the bot processes updates (messages, callbacks, etc.).
// It manages user sessions and delegates work to other handlers.
type BotHandler struct {
	BotAPI *tgApi.BotAPI
	// 所有者的 Id
	OwnerID    int64
	JobManager job.IManager
	//Sessions   map[int64]*SessionState
	Sessions smap.SyncMap[int64, *SessionState]
	Logger   zerolog.Logger
}
