package info

import (
	"context"
	"fmt"
	"strings"
	"telegram-bot/internal/config"
	"telegram-bot/internal/handlers/components"
	"telegram-bot/internal/handlers/shared"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"
)

type Handler struct{ logger *logrus.Logger }

func NewHandler(logger *logrus.Logger) *Handler { return &Handler{logger: logger} }

func (h *Handler) InfoHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !shared.IsValidPrivateMessage(update) {
		return
	}
	chatID := shared.ExtractUserID(update)
	h.logger.WithField("user_id", chatID).Info("Handler.Info.InfoHandler: sending info")
	publicSite := strings.TrimSuffix(config.Load().PublicSiteURL, "/")
	text := fmt.Sprintf("%s"+
		"Бесплатная платформа для управления записями через сайт и Telegram\n"+
		`• <a href="%s">Узнать подробнее о сервисе</a>`+
		"\n"+
		`• <a href="%s/about">Часто задаваемые вопросы</a>`+
		"\n"+
		`• <a href="%s/help">Поддержка и предложения</a>`+
		"\n"+
		"Доступные команды:\n"+
		"<blockquote>"+
		"/start — Зарегистрироваться\n"+
		"/myslots — Просмотреть свои слоты\n"+
		"/allrecords — Просмотреть историю ваших заявок\n"+
		"/myrecords — Мои предстоящие записи\n"+
		"/myrecords_confirm — Мои подтвержденные записи\n"+
		"/myrecords_reject — Мои отклоненные записи\n"+
		"/myrecords_pending — Мои записи в ожидании\n"+
		"/link — Получить свою публичную ссылку\n"+
		"/timezone — Выбрать свою таймзону\n"+
		"/upcoming — Предстоящие записи ко мне\n"+
		"</blockquote>", components.Header(), publicSite, publicSite, publicSite)
	b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, ParseMode: models.ParseModeHTML, Text: text})
}
