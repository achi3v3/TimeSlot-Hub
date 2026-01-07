package command

import (
	"strings"

	"github.com/go-telegram/bot/models"
)

const (
	command_start = "/start"
)

func IsStart(update *models.Update) bool {
	return update.Message != nil && (update.Message.Text == command_start)
}

func IsStartWithArgument(update *models.Update) bool {
	return update.Message != nil && strings.HasPrefix(update.Message.Text, command_start)
}
