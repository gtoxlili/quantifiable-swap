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

// recordMessageLog logs basic information about the incoming message.
func (handler *BotHandler) recordMessageLog(msg *tgApi.Message) {
	handler.Logger.Debug().
		Int64("UserID", msg.From.ID).
		Int64("ChatID", msg.Chat.ID).
		Str("Text", msg.Text).
		Msg("Received message")
}
