package pretty

import (
	"github.com/gtoxlili/quantifiable-swap/constants"
)

type LogData map[string]interface{}

func IsAuthorizedSubscriber(logData LogData) bool {
	subscribers, ok := logData["subscribers"].([]interface{})
	if !ok {
		return true
	}
	defer delete(logData, "subscribers")

	chatID := float64(constants.TGChatID)
	for _, subscriber := range subscribers {
		if subscriber == chatID {
			return true
		}
	}
	return false
}
