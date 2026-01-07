package timezone

import (
	"context"
	"telegram-bot/internal/handlers/shared"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"
)

type Handler struct{ logger *logrus.Logger }

func NewHandler(logger *logrus.Logger) *Handler { return &Handler{logger: logger} }

// HandlerTimezone shows timezone selection keyboard
func (h *Handler) HandlerTimezone(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !shared.IsValidPrivateMessage(update) {
		return
	}
	chatID := shared.ExtractUserID(update)
	h.logger.WithField("user_id", chatID).Info("Handler.Timezone: showing timezone selector")

	keyboard := [][]models.InlineKeyboardButton{
		{{Text: "Москва (Europe/Moscow, +03:00)", CallbackData: "tz/Europe/Moscow"}},
		{{Text: "Екатеринбург (Asia/Yekaterinburg, +05:00)", CallbackData: "tz/Asia/Yekaterinburg"}},
		{{Text: "Самара (Europe/Samara, +04:00)", CallbackData: "tz/Europe/Samara"}},
		{{Text: "Омск (Asia/Omsk, +06:00)", CallbackData: "tz/Asia/Omsk"}},
		{{Text: "Новосибирск (Asia/Novosibirsk, +07:00)", CallbackData: "tz/Asia/Novosibirsk"}},
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		ParseMode:   models.ParseModeHTML,
		Text:        "При создании слотов, клиенты будут видеть время в вашей таймзоне. Выберите пожалуйста вашу таймзону:",
		ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: keyboard},
	})
}
