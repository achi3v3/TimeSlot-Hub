package record

import (
	"context"
	"strings"
	"telegram-bot/internal/handlers/shared"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	service *Service
}

func NewHandler(b *bot.Bot, logger *logrus.Logger) *Handler {
	return &Handler{service: NewService(b, logger)}
}

func (h *Handler) HandlerGetUserRecords(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !shared.IsValidPrivateMessage(update) {
		return
	}
	userID := shared.ExtractUserID(update)
	h.service.logger.WithField("user_id", userID).Info("Handler.Record.HandlerGetUserRecords: fetching records")
	// Parse command: /my_records or /myrecords_confirm|reject|pending
	var status string
	if update.Message != nil {
		text := update.Message.Text
		if strings.Contains(text, "_") {
			parts := strings.SplitN(text, "_", 2)
			if len(parts) == 2 {
				status = parts[1]
			}
		}
	}
	h.service.SendUserRecords(ctx, userID, userID, status, 1)
}

func (h *Handler) HandlerGetAllRecords(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !shared.IsValidPrivateMessage(update) {
		return
	}
	userID := shared.ExtractUserID(update)
	h.service.logger.WithField("user_id", userID).Info("Handler.Record.HandlerGetAllRecords: fetching all records")
	h.service.SendAllRecords(ctx, userID, userID)
}
