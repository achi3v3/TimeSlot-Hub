package bot

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func AuthMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	authCache := map[int64]bool{}

	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update == nil || update.Message == nil || update.Message.From == nil {
			return
		}
		userID := update.Message.From.ID

		if authed, ok := authCache[userID]; ok && authed {
			next(ctx, b, update)
			return
		}

		responce, err := http.Get(fmt.Sprintf("http://localhost:8090/auth/users/check-auth/%d", userID))
		if err != nil || responce.StatusCode != http.StatusOK {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: userID,
				Text:   "🔒 Пожалуйста, сначала зарегистрируйтесь через /start",
			})
			return
		}

		authCache[userID] = true
		next(ctx, b, update)
	}
}
