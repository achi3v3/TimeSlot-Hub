package command

import (
	"strings"

	"github.com/go-telegram/bot/models"
)

const (
	my_slots = "/my_slots"
)

func IsGetSlots(update *models.Update) bool {
	return update.Message != nil && strings.HasPrefix(update.Message.Text, my_slots)
}
