package handler

import (
	"fmt"
	"strings"

	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// promptJobRemoval displays a list of existing jobs in a more user-friendly layout
// and asks the user to select which job they would like to remove.
func (handler *BotHandler) promptJobRemoval(chatID int64) {
	jobs := handler.JobManager.ListJobsBySubscriber(chatID)
	if len(jobs) == 0 {
		handler.sendMessage(chatID, "ℹ️ <b>系统提示</b>\n\n<i>当前没有任何可释放的任务</i>")
		return
	}

	messageText := "🗑️ <b>任务释放</b>\n\n" +
		"⚠️ <i>请谨慎操作，选择需要释放的任务</i>"

	var rows [][]tgApi.InlineKeyboardButton
	for _, jobID := range jobs {
		btnLabel := fmt.Sprintf("❌ %s", jobID.GetId())
		callbackData := fmt.Sprintf("delete_%s", jobID.GetId())
		button := tgApi.NewInlineKeyboardButtonData(btnLabel, callbackData)
		rows = append(rows, tgApi.NewInlineKeyboardRow(button))
	}
	markup := tgApi.NewInlineKeyboardMarkup(rows...)

	handler.sendMessageWithMarkup(chatID, messageText, markup)
}

func (handler *BotHandler) handleJobRemovalSelection(query *tgApi.CallbackQuery) {
	jobID := strings.TrimPrefix(query.Data, "delete_")
	chatID := query.Message.Chat.ID

	messageText := fmt.Sprintf("⚠️ <b>任务释放确认</b>\n\n"+
		"任务ID: <code>%s</code>\n"+
		"❗️<i>警告：此操作不可恢复</i>\n\n"+
		"是否继续？", jobID)
	markup := tgApi.NewInlineKeyboardMarkup(
		tgApi.NewInlineKeyboardRow(
			tgApi.NewInlineKeyboardButtonData("确定", fmt.Sprintf("confirm_delete_%s", jobID)),
			tgApi.NewInlineKeyboardButtonData("取消", "cancel_delete"),
		),
	)

	handler.sendMessageWithMarkup(chatID, messageText, markup)
}

func (handler *BotHandler) handleJobRemovalConfirmation(query *tgApi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	data := query.Data
	jobID := strings.TrimPrefix(data, "confirm_delete_")
	err := handler.JobManager.Unsubscribe(jobID, chatID)
	if err != nil {
		handler.sendEditMessage(chatID, query.Message.MessageID, fmt.Sprintf("❌ <b>释放失败</b>\n\n"+
			"任务ID: <code>%s</code>\n"+
			"错误信息: <i>%v</i>", jobID, err))
	} else {
		handler.sendEditMessage(chatID, query.Message.MessageID, fmt.Sprintf("✅ <b>释放成功</b>\n\n"+
			"任务ID: <code>%s</code>", jobID))
	}
}

// handleAdminJobRemovalConfirmation
func (handler *BotHandler) handleAdminJobRemovalConfirmation(query *tgApi.CallbackQuery) {
	jobID := strings.TrimPrefix(query.Data, "confirm_admin_delete_")
	err := handler.JobManager.DeleteJob(jobID)
	if err != nil {
		handler.sendEditMessage(query.Message.Chat.ID, query.Message.MessageID, fmt.Sprintf("❌ <b>删除失败</b>\n\n"+
			"任务ID: <code>%s</code>\n"+
			"错误信息: <i>%v</i>", jobID, err))
	} else {
		handler.sendEditMessage(query.Message.Chat.ID, query.Message.MessageID, fmt.Sprintf("✅ <b>删除成功</b>\n\n"+
			"任务ID: <code>%s</code>", jobID))
	}
}
