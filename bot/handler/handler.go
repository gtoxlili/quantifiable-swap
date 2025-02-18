package handler

import (
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gtoxlili/quantifiable-swap/common/logger"
)

func DispatchUpdates(bot *tgApi.BotAPI) {

	log := logger.NewGeneralLogger()
	u := tgApi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		log.Debug().
			Int64("User ID", update.Message.From.ID).
			Int64("Chat ID", update.Message.Chat.ID).
			Str("Text", update.Message.Text).
			Send()

		if update.Message == nil {
			msg := update.Message
			if msg.IsCommand() && msg.Chat.IsPrivate() {
				switch msg.Command() {
				case ANALYSIS:
					handleAnalysis(bot, msg.ReplyToMessage)
				}
			}
		}
	}
}

// 分析虚拟币 <币种名称> 的未来走势，基于以下信息：
//  1. 当前 RSI 为 <curRSI>，过去 RSI 列表为 <rsiQueue>。
//  2. 当前价格为 <price>，MA5 为 <ma5>，MA20 为 <ma20>。
//  3. <bar> 代表了当前 <分析级别> 周期的 K 线信息。
//  4. 以下是过去一段时间的 K 线数据（[时间，收盘价] 格式）：
//     <在此粘贴历史蜡烛数据，至少满足对应级别所需的历史数量>
//
// 请从技术面（RSI 趋势、均线支撑或压力位、价格形态等）和市场情绪方面综合分析，预测该虚拟币在接下来的若干 <分析级别> 周期内可能出现的趋势变化，并给出主要可能的上涨或下跌信号。请列出判断依据，并对可能的潜在风险或波动给出提示。
func handleAnalysis(b *tgApi.BotAPI, rm *tgApi.Message) {
	m := tgApi.NewMessage(rm.Chat.ID, "乏了，明天再说")
	m.ReplyToMessageID = rm.MessageID
	b.Send(m)
}
