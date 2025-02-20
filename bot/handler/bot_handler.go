package handler

import (
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gtoxlili/quantifiable-swap/constants"
	"github.com/gtoxlili/quantifiable-swap/job"
	"github.com/gtoxlili/quantifiable-swap/logger"
)

// NewBotHandler constructs a new BotHandler with the given bot API and job manager.
func NewBotHandler(bot *tgApi.BotAPI, manager job.IManager) *BotHandler {
	return &BotHandler{
		BotAPI:     bot,
		OwnerID:    constants.TGChatID,
		JobManager: manager,
		Logger:     logger.NewLogger("bot"),
	}
}

// StartDispatching begins listening to incoming updates and handles them.
func (handler *BotHandler) StartDispatching() {
	updateConfig := tgApi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := handler.BotAPI.GetUpdatesChan(updateConfig)
	for update := range updates {
		handler.recordMessageLog(update)
		if update.Message != nil {
			if !update.Message.Chat.IsPrivate() {
				continue
			}
			go handler.processIncomingMessage(update.Message)
		} else if update.CallbackQuery != nil {
			go handler.processCallbackQuery(update.CallbackQuery)
		}
	}
}
