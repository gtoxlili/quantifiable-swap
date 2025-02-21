package handler

import (
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
	case strings.HasPrefix(data, "manage_"):
		handler.handleJobManageSelection(query)
	case strings.HasPrefix(data, "important_only_"):
		handler.handleImportantOnlySelection(query)
	}
}

// handleConfirmation processes confirmation callbacks (e.g., job creation).
func (handler *BotHandler) handleConfirmation(query *tgApi.CallbackQuery) {
	if strings.HasSuffix(query.Data, "job_creation") {
		handler.handleJobCreationConfirmation(query)
	} else if strings.HasPrefix(query.Data, "confirm_delete_") {
		handler.handleJobRemovalConfirmation(query)
	} else if strings.HasPrefix(query.Data, "confirm_manage_") {
		handler.handleJobManageConfirmation(query)
	} else if strings.HasPrefix(query.Data, "confirm_admin_delete_") {
		handler.handleAdminJobRemovalConfirmation(query)
	}
}

func (handler *BotHandler) handleCancel(query *tgApi.CallbackQuery) {
	if strings.HasSuffix(query.Data, "job_creation") {
		handler.Sessions.Delete(query.Message.Chat.ID)
		handler.sendEditMessage(query.Message.Chat.ID, query.Message.MessageID, "🚫 <b>任务创建已取消</b>")
		return
	} else if strings.HasSuffix(query.Data, "delete") {
		handler.sendEditMessage(query.Message.Chat.ID, query.Message.MessageID, "🚫 <b>任务释放已取消</b>")
		return
	} else if strings.HasSuffix(query.Data, "manage") {
		handler.sendEditMessage(query.Message.Chat.ID, query.Message.MessageID, "🚫 <b>任务状态变更已取消</b>")
		return
	}
}
