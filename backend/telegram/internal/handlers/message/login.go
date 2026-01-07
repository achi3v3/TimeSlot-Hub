package message

import (
	"context"
	"fmt"
	"telegram-bot/internal/handlers/components"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var (
	loginMessageBase = fmt.Sprintf(
		"%s"+
			"<i>🆕 Уведомление</i>\n"+
			"<i>Чтобы подтвердить вход в аккаунт, нажмите кнопку «✔️ Подтвердить»</i>", components.Header())
)

func (h *Handler) SendLoginMessage(ctx context.Context, b *bot.Bot, userID int64, ip string, loc string) {
	msg := loginMessageBase
	// enrich with ip/location if provided
	meta := ""
	if ip != "" {
		meta = fmt.Sprintf("IP: <code>%s</code>", ip)
	}
	if loc != "" {
		if meta != "" {
			meta += " | "
		}
		meta += fmt.Sprintf("Локация: <code>%s</code>", loc)
	}
	if meta != "" {
		msg += "\n" + meta
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    userID,
		Text:      msg,
		ParseMode: models.ParseModeHTML,
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "✔️ Подтвердить", CallbackData: "ConfirmLogin"},
				},
			},
		},
	})
}
