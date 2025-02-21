package constants

import "strconv"

var (
	LogLevel       = ""
	OkxAPIKey      = ""
	OkxSecretKey   = ""
	OkxPassphrase  = ""
	ProxyAddr      = ""
	BarkToken      = ""
	ByBitAPIKey    = ""
	ByBitAPISecret = ""
	TGBotToken     = ""
	GitHubRunID    = ""
	TGChatID       int64

	tmpTgChatID = ""
)

func init() {
	TGChatID, _ = strconv.ParseInt(tmpTgChatID, 10, 64)
}
