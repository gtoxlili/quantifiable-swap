package bot

import (
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gtoxlili/quantifiable-swap/client"
	"github.com/gtoxlili/quantifiable-swap/constants"
)

var (
	Bot *tgApi.BotAPI
)

func init() {
	if constants.TGBotToken != "" {
		b, err := tgApi.NewBotAPIWithClient(constants.TGBotToken, tgApi.APIEndpoint, client.C)
		if err != nil {
			panic(err)
		}

		b.Debug = false
		Bot = b
	}
}
