package handler

import (
	"fmt"
	"strings"

	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// promptJobRemoval displays a list of existing jobs in a more user-friendly layout
// and asks the user to select which job they would like to remove.
func (handler *BotHandler) promptJobRemoval(chatID int64) {
	jobs := handler.JobManager.JobNames()
	if len(jobs) == 0 {
		handler.sendMessage(chatID, "当前没有任何可释放的任务。")
		return
	}

	messageText := "请选择要释放的任务，操作前请谨慎确认："
	var rows [][]tgApi.InlineKeyboardButton
	for _, jobID := range jobs {
		btnLabel := fmt.Sprintf("删除：%s", jobID)
		callbackData := fmt.Sprintf("delete_%s", jobID)
		button := tgApi.NewInlineKeyboardButtonData(btnLabel, callbackData)
		rows = append(rows, tgApi.NewInlineKeyboardRow(button))
	}

	markup := tgApi.NewInlineKeyboardMarkup(rows...)
	msg := tgApi.NewMessage(chatID, messageText)
	msg.ReplyMarkup = markup
	handler.BotAPI.Send(msg)
}

// handleJobRemovalSelection processes the user's choice from promptJobRemoval
// and asks for a final confirmation before deletion.
func (handler *BotHandler) handleJobRemovalSelection(query *tgApi.CallbackQuery) {
	// jobID is the job the user intends to remove.
	jobID := strings.TrimPrefix(query.Data, "delete_")
	chatID := query.Message.Chat.ID

	messageText := fmt.Sprintf("您选择删除任务 '%s'。\n此操作不可恢复，确定要删除吗？", jobID)
	msg := tgApi.NewMessage(chatID, messageText)

	markup := tgApi.NewInlineKeyboardMarkup(
		tgApi.NewInlineKeyboardRow(
			tgApi.NewInlineKeyboardButtonData("确定", fmt.Sprintf("confirm_delete_%s", jobID)),
			tgApi.NewInlineKeyboardButtonData("取消", "cancel_delete"),
		),
	)
	msg.ReplyMarkup = markup
	handler.BotAPI.Send(msg)
}

// handleJobRemovalConfirmation processes the final confirmation or cancellation.
func (handler *BotHandler) handleJobRemovalConfirmation(query *tgApi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	data := query.Data
	jobID := strings.TrimPrefix(data, "confirm_delete_")
	err := handler.JobManager.RemoveJob(jobID)
	if err != nil {
		handler.sendMessage(chatID, fmt.Sprintf("删除任务 '%s' 失败：%v", jobID, err))
	} else {
		handler.sendMessage(chatID, fmt.Sprintf("任务 '%s' 已成功删除。", jobID))
	}
}
