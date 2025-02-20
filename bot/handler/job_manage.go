package handler

import (
	"fmt"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

func (handler *BotHandler) promptJobManage(chatID int64) {
	// 超级用户才能管理任务
	if handler.OwnerID != chatID {
		handler.sendMessage(chatID, "❌ <b>权限不足</b>\n\n<i>仅管理员可管理任务</i>")
		return
	}

	jobs := handler.JobManager.ListAllJobs()
	if len(jobs) == 0 {
		handler.sendMessage(chatID, "ℹ️ <b>系统提示</b>\n\n<i>当前尚未创建任何任务</i>")
		return
	}

	messageText := "⚙️ <b>任务管理面板</b>\n\n" +
		"<i>请选择需要管理的任务：</i>"

	var rows [][]tgApi.InlineKeyboardButton
	for _, job := range jobs {
		isRunning := handler.JobManager.IsJobRunning(job.GetId())
		status := "▶️" // 暂停状态显示启动按钮
		if isRunning {
			status = "⏸️" // 运行状态显示暂停按钮
		}
		btnLabel := fmt.Sprintf("%s %s", status, job.GetId())
		callbackData := fmt.Sprintf("manage_%s", job.GetId())
		button := tgApi.NewInlineKeyboardButtonData(btnLabel, callbackData)
		rows = append(rows, tgApi.NewInlineKeyboardRow(button))
	}
	markup := tgApi.NewInlineKeyboardMarkup(rows...)

	handler.sendMessageWithMarkup(chatID, messageText, markup)
}

func (handler *BotHandler) handleJobManageSelection(query *tgApi.CallbackQuery) {
	jobID := strings.TrimPrefix(query.Data, "manage_")
	chatID := query.Message.Chat.ID

	jobData, find := handler.JobManager.GetJobData(jobID)
	if !find {
		handler.sendMessage(chatID, "❌ <b>任务不存在</b>\n\n<i>请刷新任务列表</i>")
		return
	}

	action := "暂停"
	status := "运行中"
	if !handler.JobManager.IsJobRunning(jobID) {
		action = "启动"
		status = "已暂停"
	}

	messageText := fmt.Sprintf("⚙️ <b>任务操作面板</b>\n\n"+
		"%s\n"+
		"👥 订阅者: %s\n\n"+
		"📌 当前状态: <code>%s</code>\n\n"+
		"🔔 请选择要执行的操作",
		formatJobPreview(jobData),
		formatSubscribers(handler.BotAPI, jobData.Subscribers),
		status,
	)

	markup := tgApi.NewInlineKeyboardMarkup(
		tgApi.NewInlineKeyboardRow(
			tgApi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s", action),
				fmt.Sprintf("confirm_manage_%s", jobID),
			),
			tgApi.NewInlineKeyboardButtonData("删除", fmt.Sprintf("confirm_admin_delete_%s", jobID)),
			tgApi.NewInlineKeyboardButtonData("取消", "cancel_manage"),
		),
	)

	handler.sendMessageWithMarkup(chatID, messageText, markup)
}

func (handler *BotHandler) handleJobManageConfirmation(query *tgApi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	jobID := strings.TrimPrefix(query.Data, "confirm_manage_")

	var err error
	action := "暂停"
	if !handler.JobManager.IsJobRunning(jobID) {
		err = handler.JobManager.StartJob(jobID)
		action = "启动"
	} else {
		err = handler.JobManager.StopJob(jobID)
	}

	if err != nil {
		handler.sendEditMessage(chatID, query.Message.MessageID, fmt.Sprintf("❌ <b>%s失败</b>\n\n"+
			"任务ID: <code>%s</code>\n"+
			"错误信息: <i>%v</i>", action, jobID, err))
	} else {
		handler.sendEditMessage(chatID, query.Message.MessageID, fmt.Sprintf("✅ <b>%s成功</b>\n\n"+
			"任务ID: <code>%s</code>", action, jobID))
	}
}
