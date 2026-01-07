package slot

import (
	"context"
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

func (h *Handler) HandlerGetUserSlots(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !shared.IsValidPrivateMessage(update) {
		return
	}
	userID := shared.ExtractUserID(update)
	h.service.logger.WithField("user_id", userID).Info("Handler.Slot.HandlerGetUserSlots: fetching slots")
	shared.SendGetUserSlots(ctx, b, userID, userID)
}

func (h *Handler) HandlerGetUserLink(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !shared.IsValidPrivateMessage(update) {
		return
	}
	userID := shared.ExtractUserID(update)
	h.service.logger.WithField("user_id", userID).Info("Handler.Record.HandlerGetUserRecords: fetching records")
	shared.SendGetUserLink(ctx, b, userID)
}
