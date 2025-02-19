package handler

import (
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gtoxlili/quantifiable-swap/common/job"
	"github.com/gtoxlili/quantifiable-swap/common/logger"
	"github.com/gtoxlili/quantifiable-swap/constants"
	"strconv"
)

// NewBotHandler constructs a new BotHandler with the given bot API and job manager.
func NewBotHandler(bot *tgApi.BotAPI, manager job.IManager) *BotHandler {
	ownId, _ := strconv.ParseInt(constants.TGChatID, 10, 64)
	return &BotHandler{
		BotAPI:     bot,
		OwnerID:    ownId,
		JobManager: manager,
		Sessions:   make(map[int64]*SessionState),
		Logger:     logger.NewGeneralLogger(),
	}
}

// StartDispatching begins listening to incoming updates and handles them.
func (handler *BotHandler) StartDispatching() {
	updateConfig := tgApi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := handler.BotAPI.GetUpdatesChan(updateConfig)
	for update := range updates {
		handler.recordMessageLog(update)
		// v1: 仅限 own 使用
		if update.Message != nil {
			if update.Message.Chat.ID != handler.OwnerID {
				continue
			}
			go handler.processIncomingMessage(update.Message)
		} else if update.CallbackQuery != nil {
			if update.CallbackQuery.Message.Chat.ID != handler.OwnerID {
				continue
			}
			go handler.processCallbackQuery(update.CallbackQuery)
		}
	}
}
