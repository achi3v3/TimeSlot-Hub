package formatter

import (
	"github.com/go-telegram/bot/models"
)

func IsValidPrivateMessage(update *models.Update) bool {
	return update != nil &&
		update.Message != nil &&
		update.Message.From != nil &&
		update.Message.Chat.Type == "private"
}

func ExtractUserID(update *models.Update) int64 {
	if update.Message != nil && update.Message.From != nil {
		return int64(update.Message.From.ID)
	}
	return 0
}
