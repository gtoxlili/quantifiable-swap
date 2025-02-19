package handler

import (
	"fmt"
	"strings"

	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// processCallbackQuery handles incoming callback queries from inline button interactions.
func (handler *BotHandler) processCallbackQuery(query *tgApi.CallbackQuery) {
	data := query.Data

	switch {
	case strings.HasPrefix(data, "type_"):
		handler.handleJobTypeSelection(query)
	case strings.HasPrefix(data, "delete_"):
		handler.handleJobRemovalSelection(query)
	case strings.HasPrefix(data, "confirm_"):
		handler.handleConfirmation(query)
	case strings.HasPrefix(data, "cancel_"):
		handler.handleCancel(query)

	}
}

// handleConfirmation processes confirmation callbacks (e.g., job creation).
func (handler *BotHandler) handleConfirmation(query *tgApi.CallbackQuery) {
	if strings.HasSuffix(query.Data, "job_creation") {
		session := handler.Sessions[query.Message.Chat.ID]
		if session == nil {
			handler.sendMessage(query.Message.Chat.ID, "❌ <b>创建失败</b>\n\n<i>会话已过期，请重新创建任务</i>")
			return
		}
		if _, err := handler.JobManager.AddJob(*session.TempJob); err != nil {
			handler.sendMessage(query.Message.Chat.ID, fmt.Sprintf("❌ <b>创建失败</b>\n\n错误信息：<i>%v</i>", err))
			return
		}
		handler.sendMessage(query.Message.Chat.ID, "✅ <b>任务创建成功</b>")
		delete(handler.Sessions, query.Message.Chat.ID)
	} else if strings.HasPrefix(query.Data, "confirm_delete_") {
		handler.handleJobRemovalConfirmation(query)
	}
}

func (handler *BotHandler) handleCancel(query *tgApi.CallbackQuery) {
	if strings.HasSuffix(query.Data, "job_creation") {
		delete(handler.Sessions, query.Message.Chat.ID)
		handler.sendMessage(query.Message.Chat.ID, "🚫 <b>任务创建已取消</b>")
		return
	} else if strings.HasSuffix(query.Data, "delete") {
		handler.sendMessage(query.Message.Chat.ID, "🚫 <b>任务释放已取消</b>")
		return
	}
}
