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
	case 1:
		base, quote, err := validateSymbolInput(msg.Text)
		if err != nil {
			handler.sendMessage(chatID, err.Error())
			return
		}
		session.TempJob.Symbol.Base = base
		session.TempJob.Symbol.Quote = quote
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
			handler.sendMessage(chatID, "无效的时间间隔")
			return
		}
		session.TempJob.Bar = msg.Text
		session.Step++
		handler.showJobPreview(chatID, session.TempJob)
	}
}

// showJobPreview displays a summary of the job to confirm creation.
func (handler *BotHandler) showJobPreview(chatID int64, job *config.Job) {
	msgText := fmt.Sprintf("🔄 任务预览:\n"+
		"ID: %s\n类型: %s\n交易对: %s/%s\n数量: 买 %.4f / 卖 %.4f\n"+
		"数据提供商: %s\n交易提供商: %s\n采样间隔: %s\n\n确定创建任务？",
		job.GetId(),
		job.Type,
		job.Symbol.Base,
		job.Symbol.Quote,
		job.Amount.Buy,
		job.Amount.Sell,
		job.Provider.Name,
		job.Provider.InjectOrder,
		job.Bar,
	)

	message := tgApi.NewMessage(chatID, msgText)
	keyboard := tgApi.NewInlineKeyboardMarkup(
		tgApi.NewInlineKeyboardRow(
			tgApi.NewInlineKeyboardButtonData("确定", "confirm_job_creation"),
			tgApi.NewInlineKeyboardButtonData("取消", "cancel_job_creation"),
		),
	)
	message.ReplyMarkup = keyboard
	handler.BotAPI.Send(message)
}

// promptJobType asks the user to choose a job type (e.g., SWAP or NOTIFY).
func (handler *BotHandler) promptJobType(chatID int64) {
	keyboard := tgApi.NewInlineKeyboardMarkup(
		tgApi.NewInlineKeyboardRow(
			tgApi.NewInlineKeyboardButtonData("SWAP", "type_swap"),
			tgApi.NewInlineKeyboardButtonData("NOTIFY", "type_notify"),
		),
	)

	message := tgApi.NewMessage(chatID, "请选择任务类型:")
	message.ReplyMarkup = keyboard
	handler.BotAPI.Send(message)
}

// handleJobTypeSelection captures the selected job type.
func (handler *BotHandler) handleJobTypeSelection(query *tgApi.CallbackQuery) {
	session := handler.Sessions[query.Message.Chat.ID]
	if session == nil {
		return
	}
	jobType := strings.TrimPrefix(query.Data, "type_")
	session.TempJob.Type = jobType
	session.Step++
	handler.sendMessage(query.Message.Chat.ID, "请输入交易对（格式：BASE/QUOTE）:")
}

// promptAmount asks the user for buy/sell amounts.
func (handler *BotHandler) promptAmount(chatID int64) {
	handler.sendMessage(chatID, "请输入交易数量（买入数量/卖出数量）:")
}

// promptProvider asks the user for a data or trade provider.
func (handler *BotHandler) promptProvider(title string, chatID int64) {
	handler.sendMessage(chatID, "请选择"+title+"提供商:")
}

// promptBarInterval asks the user for the data sampling interval (e.g., 15m, 1h, etc.).
func (handler *BotHandler) promptBarInterval(chatID int64) {
	handler.sendMessage(chatID, "请输入数据采样的时间间隔（如：15m, 1h, 4h, 1d）:")
}
