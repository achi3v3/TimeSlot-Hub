package message

import (
	"github.com/go-telegram/bot"
	"github.com/sirupsen/logrus"
)

const (
	command_login = "/login"
)

type Handler struct {
	bot    *bot.Bot
	logger *logrus.Logger
}

func NewHandler(b *bot.Bot, logger *logrus.Logger) *Handler {
	return &Handler{
		bot:    b,
		logger: logger,
	}
}
