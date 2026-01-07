package start

import (
	"context"
	"fmt"
	"strconv"
	"telegram-bot/internal/adapter/backendapi"
	"telegram-bot/internal/handlers/components"
	"telegram-bot/internal/handlers/shared"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"
)

const (
	command_start = "/start"
)

var (
	msgErrorWithArgument = fmt.Sprintf("%s ⚠️ Ошибка\n<i>Неверный аргумент пользователя!</i>", components.Header())
)

type Handler struct {
	service *Service
}

func NewHandler(b *bot.Bot, logger *logrus.Logger, client *backendapi.Client) *Handler {
	return &Handler{
		service: NewService(b, logger, client),
	}
}

func (h *Handler) StartHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !shared.IsValidPrivateMessage(update) {
		return
	}
	userID := shared.ExtractUserID(update)

	h.service.logger.WithField("user_id", userID).Info("Handler.Start.StartHandler: sending confirm message")
	h.service.SendConfirmMsg(ctx, b, userID)
}

func (h *Handler) StartHandlerWithArgument(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !shared.IsValidPrivateMessage(update) {
		return
	}
	userID := shared.ExtractUserID(update)
	masterID, err := strconv.ParseInt(update.Message.Text[len(command_start)+1:], 10, 0)
	if err != nil {
		h.service.logger.WithError(err).WithField("user_id", userID).Warn("Handler.Start.StartHandlerWithArgument: bad argument")
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: userID,
			Text:   msgErrorWithArgument,
		})
		return
	}
	h.service.logger.WithFields(logrus.Fields{"user_id": userID, "master_id": masterID}).Info("Handler.Start.StartHandlerWithArgument: fetching slots")
	shared.SendFutureSlotsForClient(ctx, b, userID, masterID, 1, 0)
}
