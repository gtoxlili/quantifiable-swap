package handler

import (
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// processIncomingMessage inspects the user's session state (if any) and
// delegates either stateful flow handling or command handling.
func (handler *BotHandler) processIncomingMessage(msg *tgApi.Message) {
	session := handler.Sessions[msg.Chat.ID]
	switch {
	case session != nil:
		handler.handleStatefulFlow(msg, session)
	case msg.IsCommand():
		handler.handleCommand(msg)
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
	case "addjob":
		handler.initializeJobCreation(chatID)
		handler.promptJobType(chatID)
	case "removejob":
		handler.promptJobRemoval(chatID)
	}
}

// recordMessageLog logs detailed information about the incoming update.
func (handler *BotHandler) recordMessageLog(u tgApi.Update) {
	log := handler.Logger.Debug()

	if u.Message != nil {
		log.Int64("UserID", u.Message.From.ID).
			Int64("ChatID", u.Message.Chat.ID).
			Str("Text", u.Message.Text).
			Str("Username", u.Message.From.UserName).
			Int("MessageID", u.Message.MessageID).
			Str("Type", "message")

		if u.Message.IsCommand() {
			log.Str("Command", u.Message.Command()).
				Str("CommandArgs", u.Message.CommandArguments())
		}

		if u.Message.ReplyToMessage != nil {
			log.Int("ReplyToMessageID", u.Message.ReplyToMessage.MessageID).
				Int64("ReplyToUserID", u.Message.ReplyToMessage.From.ID).
				Str("ReplyToUsername", u.Message.ReplyToMessage.From.UserName).
				Str("ReplyToText", u.Message.ReplyToMessage.Text)
		}
	}

	if u.CallbackQuery != nil {
		log.Int64("UserID", u.CallbackQuery.From.ID).
			Int64("ChatID", u.CallbackQuery.Message.Chat.ID).
			Str("CallbackData", u.CallbackQuery.Data).
			Str("Username", u.CallbackQuery.From.UserName).
			Int("MessageID", u.CallbackQuery.Message.MessageID).
			Str("Type", "callback")
	}

	log.Send()
}
