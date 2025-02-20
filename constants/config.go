package constants

import "strconv"

var (
	OkxAPIKey      = ""
	OkxSecretKey   = ""
	OkxPassphrase  = ""
	ProxyAddr      = ""
	BarkToken      = ""
	ByBitAPIKey    = ""
	ByBitAPISecret = ""
	TGBotToken     = ""
	TGChatID       int64

	tmpTgChatID = ""
)

func init() {
	TGChatID, _ = strconv.ParseInt(tmpTgChatID, 10, 64)
}
