package handler

import (
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/common/config"
	"strings"
	"time"

	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// initializeJobCreation sets up a new session state for creating a job.
func (handler *BotHandler) initializeJobCreation(chatID int64) {
	handler.Sessions[chatID] = &SessionState{
		CurrentAction: "action_create_job",
		TempJob:       &config.Job{},
		Step:          0,
	}
}

// continueJobCreation advances the job creation flow based on the current step.
func (handler *BotHandler) continueJobCreation(msg *tgApi.Message, session *SessionState) {
	chatID := msg.Chat.ID

	switch session.Step {
	case 0:
		handler.promptJobType(chatID)
	case 1:
		base, quote, err := validateSymbolInput(msg.Text)
		if err != nil {
			handler.sendMessage(chatID, err.Error())
			return
		}
		session.TempJob.Symbol.Base = base
		session.TempJob.Symbol.Quote = quote
		if session.TempJob.Type == "notify" {
			session.Step = 3
			handler.promptProvider("数据", chatID)
			return
		}
		session.Step++
		handler.promptAmount(chatID)
	case 2:
		buyAm, sellAm, err := validateAmountInput(msg.Text)
		if err != nil {
			handler.sendMessage(chatID, err.Error())
			return
		}
		session.TempJob.Amount.Buy = buyAm
		session.TempJob.Amount.Sell = sellAm
		session.Step++
		handler.promptProvider("数据", chatID)
	case 3:
		session.TempJob.Provider.Name = strings.TrimSpace(msg.Text)
		if session.TempJob.Type == "notify" {
			session.Step = 5
			handler.promptBarInterval(chatID)
			return
		}
		session.Step++
		handler.promptProvider("交易", chatID)
	case 4:
		if msg.Text != session.TempJob.Provider.Name {
			session.TempJob.Provider.InjectOrder = strings.TrimSpace(msg.Text)
		}
		session.Step++
		handler.promptBarInterval(chatID)
	case 5:
		if _, err := time.ParseDuration(msg.Text); err != nil {
			handler.sendMessage(chatID, "❌ <b>时间格式错误</b>\n\n<i>请输入有效的时间间隔</i>")
			return
		}
		session.TempJob.Bar = msg.Text
		session.Step++
		handler.showJobPreview(chatID, session.TempJob)
	}
}

// showJobPreview displays a summary of the job to confirm creation.
func (handler *BotHandler) showJobPreview(chatID int64, job *config.Job) {
	var msgText string
	if job.Type == "notify" {
		msgText = fmt.Sprintf(
			"📋 <b>任务预览</b>\n\n"+
				"🔑 ID: <code>%s</code>\n"+
				"📊 类型: <code>%s</code>\n"+
				"💱 交易对: <code>%s/%s</code>\n"+
				"📡 数据提供商: <code>%s</code>\n"+
				"⏱️ 采样间隔: <code>%s</code>\n\n"+
				"⚠️ <i>请确认以上信息</i>",
			job.GetId(), job.Type, job.Symbol.Base, job.Symbol.Quote,
			job.Provider.Name, job.Bar,
		)
	} else {
		ordPb := job.Provider.InjectOrder
		if ordPb == "" {
			ordPb = job.Provider.Name
		}
		msgText = fmt.Sprintf(
			"📋 <b>任务预览</b>\n\n"+
				"🔑 ID: <code>%s</code>\n"+
				"📊 类型: <code>%s</code>\n"+
				"💱 交易对: <code>%s/%s</code>\n"+
				"💰 数量: 买入 <code>%.4f</code> / 卖出 <code>%.4f</code>\n"+
				"📡 数据提供商: <code>%s</code>\n"+
				"🏛️ 交易提供商: <code>%s</code>\n"+
				"⏱️ 采样间隔: <code>%s</code>\n\n"+
				"⚠️ <i>请确认以上信息</i>",
			job.GetId(), job.Type, job.Symbol.Base, job.Symbol.Quote,
			job.Amount.Buy, job.Amount.Sell, job.Provider.Name,
			ordPb, job.Bar,
		)
	}

	keyboard := tgApi.NewInlineKeyboardMarkup(
		tgApi.NewInlineKeyboardRow(
			tgApi.NewInlineKeyboardButtonData("确定", "confirm_job_creation"),
			tgApi.NewInlineKeyboardButtonData("取消", "cancel_job_creation"),
		),
	)
	handler.sendMessageWithMarkup(chatID, msgText, keyboard)
}

func (handler *BotHandler) promptJobType(chatID int64) {
	keyboard := tgApi.NewInlineKeyboardMarkup(
		tgApi.NewInlineKeyboardRow(
			tgApi.NewInlineKeyboardButtonData("SWAP", "type_swap"),
			tgApi.NewInlineKeyboardButtonData("NOTIFY", "type_notify"),
		),
	)
	handler.sendMessageWithMarkup(chatID, "📝 <b>选择任务类型</b>", keyboard)
}

func (handler *BotHandler) handleJobTypeSelection(query *tgApi.CallbackQuery) {
	session := handler.Sessions[query.Message.Chat.ID]
	if session == nil {
		return
	}
	jobType := strings.TrimPrefix(query.Data, "type_")
	session.TempJob.Type = jobType
	session.Step++
	handler.sendMessage(query.Message.Chat.ID, "💱 <b>输入交易对</b>\n\n格式：<code>BASE/QUOTE</code>")
}

func (handler *BotHandler) promptAmount(chatID int64) {
	handler.sendMessage(chatID, "💰 <b>输入交易数量</b>\n\n格式：<code>买入数量/卖出数量</code>")
}

func (handler *BotHandler) promptProvider(title string, chatID int64) {
	icon := map[string]string{
		"数据": "📡",
		"交易": "🏛️",
	}[title]
	handler.sendMessage(chatID, fmt.Sprintf("%s <b>选择%s提供商</b>", icon, title))
}

func (handler *BotHandler) promptBarInterval(chatID int64) {
	handler.sendMessage(chatID, "⏱️ <b>设置采样间隔</b>\n\n"+
		"支持的格式：\n"+
		"• <code>15m</code> - 15分钟\n"+
		"• <code>1h</code>  - 1小时\n"+
		"• <code>4h</code>  - 4小时\n"+
		"• <code>1d</code>  - 1天")
}
