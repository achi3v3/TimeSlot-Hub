package bot

import (
	"context"
	"telegram-bot/internal/adapter/backendapi"
	appLogin "telegram-bot/internal/app/login"
	appRecords "telegram-bot/internal/app/record"
	appSlots "telegram-bot/internal/app/slots"
	botMiddleware "telegram-bot/internal/bot"
	"telegram-bot/internal/config"
	hInfo "telegram-bot/internal/handlers/info"
	hMaster "telegram-bot/internal/handlers/master"
	hRecord "telegram-bot/internal/handlers/record"
	hSlot "telegram-bot/internal/handlers/slot"
	hStart "telegram-bot/internal/handlers/start"
	hTimezone "telegram-bot/internal/handlers/timezone"
	"time"

	"github.com/go-telegram/bot"
	"github.com/sirupsen/logrus"
)

type Server struct {
	bot      *bot.Bot
	logger   *logrus.Logger
	login    *appLogin.Service
	slots    *appSlots.Service
	records  *appRecords.Service
	stopFunc context.CancelFunc
	name     string
}

func NewServer(b *bot.Bot, logger *logrus.Logger, login *appLogin.Service, slots *appSlots.Service, records *appRecords.Service, name string) *Server {
	return &Server{
		bot:     b,
		logger:  logger,
		login:   login,
		slots:   slots,
		records: records,
		name:    name,
	}
}

func (s *Server) RegisterHandlers() {
	cfg := config.Load()
	client := backendapi.New(cfg.BackendBaseURL, logrus.New())
	s.logger.Infof("Transport.Bot.Server.RegisterHandlers: begin")
	startHandler := hStart.NewHandler(s.bot, s.logger, client)
	slotHandler := hSlot.NewHandler(s.bot, s.logger)
	recordHandler := hRecord.NewHandler(s.bot, s.logger)
	infoHandler := hInfo.NewHandler(s.logger)
	timezoneHandler := hTimezone.NewHandler(s.logger)
	masterHandler := hMaster.NewHandler(s.logger)

	// Применяем rate limiting middleware к командам
	s.bot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, botMiddleware.CommandRateLimitMiddleware(startHandler.StartHandler))
	s.bot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypePrefix, botMiddleware.CommandRateLimitMiddleware(startHandler.StartHandlerWithArgument))
	s.bot.RegisterHandler(bot.HandlerTypeMessageText, "/myslots", bot.MatchTypeExact, botMiddleware.CommandAuthMiddleware(botMiddleware.CommandRateLimitMiddleware(slotHandler.HandlerGetUserSlots), client))
	s.bot.RegisterHandler(bot.HandlerTypeMessageText, "/allrecords", bot.MatchTypeExact, botMiddleware.CommandAuthMiddleware(botMiddleware.CommandRateLimitMiddleware(recordHandler.HandlerGetAllRecords), client))
	s.bot.RegisterHandler(bot.HandlerTypeMessageText, "/myrecords", bot.MatchTypePrefix, botMiddleware.CommandAuthMiddleware(botMiddleware.CommandRateLimitMiddleware(recordHandler.HandlerGetUserRecords), client))
	s.bot.RegisterHandler(bot.HandlerTypeMessageText, "/link", bot.MatchTypeExact, botMiddleware.CommandAuthMiddleware(botMiddleware.CommandRateLimitMiddleware(slotHandler.HandlerGetUserLink), client))
	s.bot.RegisterHandler(bot.HandlerTypeMessageText, "/info", bot.MatchTypeExact, botMiddleware.CommandRateLimitMiddleware(infoHandler.InfoHandler))
	s.bot.RegisterHandler(bot.HandlerTypeMessageText, "/upcoming", bot.MatchTypeExact, botMiddleware.CommandAuthMiddleware(botMiddleware.CommandRateLimitMiddleware(masterHandler.HandlerUpcomingRecords), client))
	// /timezone requires auth, rate-limited
	s.bot.RegisterHandler(bot.HandlerTypeMessageText, "/timezone", bot.MatchTypeExact, botMiddleware.CommandAuthMiddleware(botMiddleware.CommandRateLimitMiddleware(timezoneHandler.HandlerTimezone), client))

	s.logger.Infof("Transport.Bot.Server.RegisterHandlers: registered handlers with rate limiting")
}

func (s *Server) Start(ctx context.Context) error {
	ctx, s.stopFunc = context.WithCancel(ctx)

	s.logger.Infof("Transport.Bot.Server.Start: polling start")
	s.bot.Start(ctx)
	return nil
}
func (s *Server) Shutdown(ctx context.Context) error {
	if s.stopFunc != nil {
		s.stopFunc()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Second):
	}
	s.logger.Infof("%s shutdown complete", s.name)
	return nil
}

func (s *Server) Name() string { return s.name }
