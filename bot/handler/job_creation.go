package handler

import (
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/common/config"
	"github.com/gtoxlili/quantifiable-swap/market"
	"strings"
	"time"

	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// initializeJobCreation sets up a new session state for creating a job.
func (handler *BotHandler) initializeJobCreation(chatID int64) {
	handler.Sessions.Store(chatID, &SessionState{
		CurrentAction: "action_create_job",
		TempJob:       &config.Job{Subscribers: []config.Subscriber{{ID: chatID}}},
		Step:          0,
	})
}

// 游客只能创建 monitor 类型的任务
func (handler *BotHandler) initializeJobCreationForGuest(chatID int64) {
	handler.Sessions.Store(chatID, &SessionState{
		CurrentAction: "action_create_job",
		TempJob:       &config.Job{Type: "monitor", Subscribers: []config.Subscriber{{ID: chatID}}},
		Step:          1,
	})
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
		if session.TempJob.Type == "monitor" {
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
		prov, err := validateProviderInput("数据", msg.Text)
		if err != nil {
			handler.sendMessage(chatID, err.Error())
			return
		}
		session.TempJob.Provider.Data = prov
		if session.TempJob.Type == "monitor" {
			session.Step = 5
			handler.promptBarInterval(chatID)
			return
		}
		session.Step++
		handler.promptProvider("交易", chatID)
	case 4:
		prov, err := validateProviderInput("交易", msg.Text)
		if err != nil {
			handler.sendMessage(chatID, err.Error())
			return
		}
		session.TempJob.Provider.Trading = prov
		session.Step++
		handler.promptBarInterval(chatID)
	case 5:
		if _, err := time.ParseDuration(msg.Text); err != nil {
			handler.sendMessage(chatID, "❌ <b>时间格式错误</b>\n\n<i>请输入有效的时间间隔</i>")
			return
		}
		session.TempJob.Bar = msg.Text
		session.Step++
		handler.promptImportantOnly(chatID)
	case 6:
		handler.promptImportantOnly(chatID)
	}
}

// showJobPreview displays a summary of the job to confirm creation.
func (handler *BotHandler) showJobPreview(chatID int64, job config.Job) {
	keyboard := tgApi.NewInlineKeyboardMarkup(
		tgApi.NewInlineKeyboardRow(
			tgApi.NewInlineKeyboardButtonData("确定", "confirm_job_creation"),
			tgApi.NewInlineKeyboardButtonData("取消", "cancel_job_creation"),
		),
	)
	notifyLevel := "仅重要通知"
	if !job.Subscribers[0].ImportantOnly {
		notifyLevel = "全部通知"
	}
	handler.sendMessageWithMarkup(chatID,
		fmt.Sprintf("📋 <b>任务预览</b>\n\n%s\n🔔 通知级别: <code>%s</code>\n\n⚠️ <i>请确认以上信息</i>", formatJobPreview(job), notifyLevel),
		keyboard)
}

func (handler *BotHandler) promptJobType(chatID int64) {
	keyboard := tgApi.NewInlineKeyboardMarkup(
		tgApi.NewInlineKeyboardRow(
			tgApi.NewInlineKeyboardButtonData("TRADER", "type_trader"),
		),
		tgApi.NewInlineKeyboardRow(
			tgApi.NewInlineKeyboardButtonData("MONITOR", "type_monitor"),
		),
	)
	handler.sendMessageWithMarkup(chatID, "📝 <b>选择任务类型</b>", keyboard)
}

func (handler *BotHandler) handleJobTypeSelection(query *tgApi.CallbackQuery) {
	session, found := handler.Sessions.Load(query.Message.Chat.ID)
	if !found {
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

	providers := market.ListAvailableProviders(title)
	var providerList strings.Builder
	for _, p := range providers {
		providerList.WriteString(fmt.Sprintf("• <code>%s</code>\n", p))
	}

	handler.sendMessage(chatID, fmt.Sprintf("%s <b>选择%s提供商</b>\n\n"+
		"支持的提供商：\n%s", icon, title, providerList.String()))
}

func (handler *BotHandler) promptBarInterval(chatID int64) {
	handler.sendMessage(chatID, "⏱️ <b>设置采样间隔</b>\n\n"+
		"支持的格式：\n"+
		"• <code>15m</code> - 15分钟\n"+
		"• <code>1h</code>  - 1小时\n"+
		"• <code>4h</code>  - 4小时\n"+
		"• <code>1d</code>  - 1天")
}

func (handler *BotHandler) promptImportantOnly(chatID int64) {
	keyboard := tgApi.NewInlineKeyboardMarkup(
		tgApi.NewInlineKeyboardRow(
			tgApi.NewInlineKeyboardButtonData("是", "important_only_yes"),
			tgApi.NewInlineKeyboardButtonData("否", "important_only_no"),
		),
	)
	handler.sendMessageWithMarkup(chatID, "🔔 <b>是否只接收重要通知</b>\n\n"+
		"• 重要通知包含：交易信号、错误警告\n"+
		"• 普通通知包含：市场数据、状态更新", keyboard)
}

// handleJobCreationConfirmation processes the user's confirmation to create a job.
func (handler *BotHandler) handleJobCreationConfirmation(query *tgApi.CallbackQuery) {
	session, ok := handler.Sessions.Load(query.Message.Chat.ID)
	if !ok {
		handler.sendEditMessage(query.Message.Chat.ID, query.Message.MessageID, "❌ <b>创建失败</b>\n\n<i>会话已过期，请重新创建任务</i>")
		return
	}
	if _, err := handler.JobManager.CreateJob(*session.TempJob); err != nil {
		handler.sendEditMessage(query.Message.Chat.ID, query.Message.MessageID, fmt.Sprintf("❌ <b>创建失败</b>\n\n错误信息：<i>%v</i>", err))
		return
	}
	_ = handler.JobManager.StartJob(session.TempJob.String())
	handler.sendEditMessage(query.Message.Chat.ID, query.Message.MessageID, "✅ <b>任务创建成功</b>")
	handler.Sessions.Delete(query.Message.Chat.ID)
}

// handleImportantOnlySelection
func (handler *BotHandler) handleImportantOnlySelection(query *tgApi.CallbackQuery) {
	session, ok := handler.Sessions.Load(query.Message.Chat.ID)
	if !ok {
		handler.sendEditMessage(query.Message.Chat.ID, query.Message.MessageID, "❌ <b>创建失败</b>\n\n<i>会话已过期，请重新创建任务</i>")
		return
	}
	if query.Data == "important_only_yes" {
		session.TempJob.Subscribers[0].ImportantOnly = true
	}
	session.Step++
	handler.showJobPreview(query.Message.Chat.ID, *session.TempJob)
}
