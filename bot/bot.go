package bot

import (
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gtoxlili/quantifiable-swap/client"
	"github.com/gtoxlili/quantifiable-swap/constants"
)

var (
	B *Bot
)

type Bot struct {
	*tgApi.BotAPI
}

func New(token string) *Bot {
	bot, err := tgApi.NewBotAPIWithClient(token, tgApi.APIEndpoint, client.C)
	if err != nil {
		panic(err)
	}

	bot.Debug = false

	return &Bot{bot}
}

func init() {
	if constants.TGBotToken != "" {
		B = New(constants.TGBotToken)
	}
}
