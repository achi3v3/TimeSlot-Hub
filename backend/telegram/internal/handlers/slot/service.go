package slot

import (
	"github.com/go-telegram/bot"
	"github.com/sirupsen/logrus"
)

type Service struct {
	bot    *bot.Bot
	logger *logrus.Logger
}

func NewService(bot *bot.Bot, logger *logrus.Logger) *Service {
	return &Service{bot: bot, logger: logger}
}
