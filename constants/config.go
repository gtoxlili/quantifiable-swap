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
	TGBotToken     = "7102299208:AAHiPGCMtPWvMYwrm62eW_bpkperpsPncMg"
	TGChatID       int64

	tmpTgChatID = "584544685"
)

func init() {
	TGChatID, _ = strconv.ParseInt(tmpTgChatID, 10, 64)
}
