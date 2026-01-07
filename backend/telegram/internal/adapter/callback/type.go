package callback

import (
	"context"
	adapter "telegram-bot/internal/adapter/backendapi"
	"telegram-bot/internal/config"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type CallBackHandler struct {
	Handler
	messageID int
	userID    int64
	query     string
}

func NewCallBackHandler(handler Handler,
	messageID int,
	userID int64,
	query string,
) *CallBackHandler {
	return &CallBackHandler{
		Handler:   handler,
		messageID: messageID,
		userID:    userID,
		query:     query,
	}
}

type Handler struct {
	client adapter.Client
	config config.Config
	ctx    context.Context
	b      *bot.Bot
	update *models.Update
}

func NewHandler(client adapter.Client,
	config config.Config,
	ctx context.Context,
	b *bot.Bot,
	update *models.Update) *Handler {
	return &Handler{
		client: client,
		config: config,
		ctx:    ctx,
		b:      b,
		update: update,
	}
}
