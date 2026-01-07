package start

import (
	"context"
	"fmt"
	"strings"
	"telegram-bot/internal/adapter/backendapi"
	"telegram-bot/internal/config"
	"telegram-bot/internal/handlers/components"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"
)

var (
	cfg          = config.Load()
	publicSite   = strings.TrimSuffix(cfg.PublicSiteURL, "/")
	startMessage = fmt.Sprintf(
		"%s"+
			"💬 Привет!\n"+
			"Для дальнейшей работы с ботом, требуется чтобы вы зарегистрировались, нажав кнопку «Подтвердить».\n"+
			"\n<blockquote><b>Нажав, вы делитесь контактом и подтверждаете:</b>\n"+
			`• <a href="%s/privacy">Согласие на обработку персональных данных</a>`+
			"\n"+
			`• <a href="%s/terms">Согласие с Пользовательским соглашением</a>`+
			"\n"+
			"• Что ваш номер будет использоваться для входа и идентификации мастером вас, как пользователя\n</blockquote>"+
			"\n"+
			"\nАккаунт можно будет в любое время удалить, на сайте в самом низу вкладки «Профиль»\n"+
			`<a href="%s/about">Узнать больше о нас</a>`+
			"\n", components.Header(), publicSite, publicSite, publicSite)
)

type Service struct {
	bot    *bot.Bot
	client *backendapi.Client
	logger *logrus.Logger
}

func NewService(bot *bot.Bot, logger *logrus.Logger, client *backendapi.Client) *Service {
	return &Service{
		bot:    bot,
		client: client,
		logger: logger,
	}
}

func (s *Service) SendConfirmMsg(ctx context.Context, b *bot.Bot, userID int64) {

	_, exist := s.client.CheckAuth(ctx, userID)
	if exist {
		msgText := fmt.Sprintf(
			"%s"+
				"<blockquote> ℹ️ Вы уже зарегистрированы!\n\n<b>Повторная регистрация не требуется</b> </blockquote>",
			components.Header(),
		)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    userID,
			ParseMode: models.ParseModeHTML,
			Text:      msgText,
		})
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    userID,
		Text:      startMessage,
		ParseMode: models.ParseModeHTML,
		ReplyMarkup: &models.ReplyKeyboardMarkup{
			Keyboard: [][]models.KeyboardButton{
				{
					{Text: "✔️ Подтвердить", RequestContact: true},
				},
			},
		},
	})
}
