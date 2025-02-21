package handler

import (
	"fmt"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gtoxlili/quantifiable-swap/constants"
)

// processIncomingMessage inspects the user's session state (if any) and
// delegates either stateful flow handling or command handling.
func (handler *BotHandler) processIncomingMessage(msg *tgApi.Message) {
	session, _ := handler.Sessions.Load(msg.Chat.ID)
	switch {
	case msg.IsCommand():
		handler.handleCommand(msg)
	case session != nil:
		handler.handleStatefulFlow(msg, session)
	}
}

// handleStatefulFlow routes the message to the relevant step based on the session state.
func (handler *BotHandler) handleStatefulFlow(msg *tgApi.Message, session *SessionState) {
	switch session.CurrentAction {
	case "action_create_job":
		handler.continueJobCreation(msg, session)
		// Add other stateful flows here as needed.
	}
}

// handleCommand processes bot commands.
func (handler *BotHandler) handleCommand(msg *tgApi.Message) {
	chatID := msg.Chat.ID
	cmd := msg.Command()
	switch cmd {
	case "add":
		if handler.OwnerID != chatID {
			handler.initializeJobCreationForGuest(chatID)
			handler.sendMessage(chatID, "💱 <b>输入交易对</b>\n\n格式：<code>BASE/QUOTE</code>")
			return
		}
		handler.initializeJobCreation(chatID)
		handler.promptJobType(chatID)
	case "remove":
		handler.promptJobRemoval(chatID)
	// 管理 （暂停或者启动）
	case "manage":
		handler.promptJobManage(chatID)
	case "version":
		handler.promptVersionInfo(constants.GitHubRunID)
	}
}

// recordMessageLog logs detailed information about the incoming update.
func (handler *BotHandler) recordMessageLog(u tgApi.Update) {
	log := handler.Logger.Debug()

	if u.Message != nil {
		if u.Message.From.IsBot {
			return
		}
		log.Str("UserID", fmt.Sprintf("%d", u.Message.From.ID)).
			Str("ChatID", fmt.Sprintf("%d", u.Message.Chat.ID)).
			Str("Text", u.Message.Text).
			Str("Username", u.Message.From.UserName).
			Str("MessageID", fmt.Sprintf("%d", u.Message.MessageID)).
			Str("Type", "message")

		if u.Message.IsCommand() {
			log.Str("Command", u.Message.Command()).
				Str("CommandArgs", u.Message.CommandArguments())
		}

		if u.Message.ReplyToMessage != nil {
			log.Str("ReplyToMessageID", fmt.Sprintf("%d", u.Message.ReplyToMessage.MessageID)).
				Str("ReplyToUserID", fmt.Sprintf("%d", u.Message.ReplyToMessage.From.ID)).
				Str("ReplyToUsername", u.Message.ReplyToMessage.From.UserName).
				Str("ReplyToText", u.Message.ReplyToMessage.Text)
		}
	}

	if u.CallbackQuery != nil {
		log.Str("UserID", fmt.Sprintf("%d", u.CallbackQuery.From.ID)).
			Str("ChatID", fmt.Sprintf("%d", u.CallbackQuery.Message.Chat.ID)).
			Str("CallbackData", u.CallbackQuery.Data).
			Str("Username", u.CallbackQuery.From.UserName).
			Str("MessageID", fmt.Sprintf("%d", u.CallbackQuery.Message.MessageID)).
			Str("Type", "callback")
	}

	log.Send()
}

func (handler *BotHandler) promptVersionInfo(runID string) {
	messageText := fmt.Sprintf("🎯<b>Github Run Id: <a href='https://github.com/gtoxlili/quantifiable-swap/actions/runs/%s'>%s</a></b>\n\n"+
		"🚀 Quantifiable Swap 由 GitHub Actions 构建并部署至云服务器\n"+
		"✅ 此信息可帮助你确认当前版本是由开源代码所构建\n"+
		"⚠️ 未经任何人（包含作者）修改",
		runID, runID)
	handler.sendMessage(handler.OwnerID, messageText)
}
