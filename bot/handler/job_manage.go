package handler

import (
	"fmt"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gtoxlili/quantifiable-swap/common/config"
	"github.com/gtoxlili/quantifiable-swap/common/lo"
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
		isRunning := handler.JobManager.IsJobRunning(job.String())
		status := "▶️" // 暂停状态显示启动按钮
		if isRunning {
			status = "⏸️" // 运行状态显示暂停按钮
		}
		btnLabel := fmt.Sprintf("%s %s", status, job.String())
		callbackData := fmt.Sprintf("manage_%s", job.String())
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
		formatSubscribers(handler.BotAPI, lo.Map(jobData.Subscribers, func(sub config.Subscriber) int64 {
			return sub.ID
		})),
		status,
	)

	markup := tgApi.NewInlineKeyboardMarkup(
		tgApi.NewInlineKeyboardRow(
			tgApi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s", action),
				fmt.Sprintf("confirm_manage_%s", jobID),
			),
			tgApi.NewInlineKeyboardButtonData("编辑", fmt.Sprintf("edit_%s", jobID)),
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

func (handler *BotHandler) initializeJobEdit(chatID int64, jobData *config.Job) {
	handler.Sessions.Store(chatID, &SessionState{
		CurrentAction: "action_edit_job",
		TempJob:       jobData,
		Step:          0,
	})
}

func (handler *BotHandler) handleJobEditSelection(query *tgApi.CallbackQuery) {
	jobID := strings.TrimPrefix(query.Data, "edit_")
	chatID := query.Message.Chat.ID

	jobData, find := handler.JobManager.GetJobData(jobID)
	if !find {
		handler.sendMessage(chatID, "❌ <b>任务不存在</b>\n\n<i>请刷新任务列表</i>")
		return
	}

	handler.initializeJobEdit(chatID, &jobData)
	handler.sendEditMessage(chatID, query.Message.MessageID,
		fmt.Sprintf("✏️ <b>编辑任务配置</b>\n\n"+
			"当前配置：\n"+
			"<pre><code class=\"language-json\">%s</code></pre>\n\n"+
			"请回复修改后的 JSON 配置：\n"+
			"• 保持 JSON 格式\n"+
			"• 可直接粘贴完整 JSON 或只修改部分字段",
			jobData.Format()))
}

// continueJobEdit
func (handler *BotHandler) continueJobEdit(msg *tgApi.Message, session *SessionState) {
	chatID := msg.Chat.ID

	switch session.Step {
	case 0:
		editJob, err := validateJobInput(msg.Text)
		if err != nil {
			handler.sendMessage(chatID, err.Error())
			return
		}
		defer handler.deleteMessage(chatID, msg.MessageID)
		session.Step++
		session.TempEditJob = editJob
		// 任务修改预览
		keyboard := tgApi.NewInlineKeyboardMarkup(
			tgApi.NewInlineKeyboardRow(
				tgApi.NewInlineKeyboardButtonData("确定", "confirm_job_edit"),
				tgApi.NewInlineKeyboardButtonData("取消", "cancel_job_edit"),
			),
		)
		handler.sendMessageWithMarkup(chatID,
			fmt.Sprintf("📋 <b>任务编辑预览</b>\n\n<pre><code class=\"language-json\">%s</code></pre>\n\n⚠️ <i>请确认以上信息</i>", editJob.Format()),
			keyboard)
	}
}

func (handler *BotHandler) handleJobEditConfirmation(query *tgApi.CallbackQuery) {
	session, found := handler.Sessions.Load(query.Message.Chat.ID)
	if !found {
		handler.sendEditMessage(query.Message.Chat.ID, query.Message.MessageID,
			"❌ <b>会话异常</b>\n\n"+
				"• 状态：<i>会话已过期</i>\n"+
				"• 建议：<i>请重新编辑任务</i>")
		return
	}

	// 删除旧任务
	if err := handler.JobManager.DeleteJob(session.TempJob.String()); err != nil {
		handler.sendEditMessage(query.Message.Chat.ID, query.Message.MessageID,
			fmt.Sprintf("❌ <b>任务更新失败</b>\n\n"+
				"• 任务ID：<code>%s</code>\n"+
				"• 错误信息：<i>%v</i>\n"+
				"• 失败阶段：<i>删除旧任务</i>",
				session.TempJob.String(), err))
		return
	}

	// 创建新任务
	if _, err := handler.JobManager.CreateJob(*session.TempEditJob); err != nil {
		handler.sendEditMessage(query.Message.Chat.ID, query.Message.MessageID,
			fmt.Sprintf("❌ <b>任务更新失败</b>\n\n"+
				"• 任务ID：<code>%s</code>\n"+
				"• 错误信息：<i>%v</i>\n"+
				"• 失败阶段：<i>创建新任务</i>",
				session.TempJob.String(), err))
		return
	}

	_ = handler.JobManager.StartJob(session.TempEditJob.String())
	handler.sendEditMessage(query.Message.Chat.ID, query.Message.MessageID,
		"✅ <b>任务更新成功</b>\n\n"+
			"• 状态：<i>已自动启动</i>")
	handler.Sessions.Delete(query.Message.Chat.ID)
}
