package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	adapter "telegram-bot/internal/adapter/backendapi"
	callback "telegram-bot/internal/adapter/callback"
	appHandler "telegram-bot/internal/app/login"
	appRecords "telegram-bot/internal/app/record"
	appSlots "telegram-bot/internal/app/slots"
	mybot "telegram-bot/internal/bot"
	"telegram-bot/internal/config"
	msgHandler "telegram-bot/internal/handlers/message"
	"telegram-bot/internal/logger"
	botServer "telegram-bot/internal/transport/bot"
	httpTransport "telegram-bot/internal/transport/http"
	"telegram-bot/pkg/closer"
	"time"

	"github.com/go-telegram/bot"
)

// LaunchBot starts Telegram bot and HTTP server
func LaunchBot() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg := config.Load()
	log := logger.New()
	log.Infof("cmd.launchBot: starting with base_url=%s", cfg.BackendBaseURL)

	if cfg.BotToken == "" {
		log.Error("cmd.launchBot: BOT_TOKEN environment variable is not set")
		return
	}
	if cfg.BackendBaseURL == "" {
		log.Fatal("cmd.launchBot: BACKEND_BASE_URL environment variable is required")
		return
	}

	mybot.InitRateLimiter()
	closeRateLimiter := mybot.NewCloseRateLimiter("close-rate-limiter")

	log.Info("cmd.launchBot: rate limiter initialized")

	opts := []bot.Option{
		bot.WithDefaultHandler(callback.UniversalHandler),
	}
	b, err := bot.New(cfg.BotToken, opts...)
	if err != nil {
		log.WithError(err).Error("cmd.launchBot: failed to create telegram bot instance")
		return
	}
	log.Info("cmd.launchBot: telegram bot instance created successfully")

	stateManager := msgHandler.GetStateManager()
	stateManager.StartCleanupRoutine()
	log.Info("cmd.launchBot: message state manager initialized and cleanup routine started")

	apiClient := adapter.New(cfg.BackendBaseURL, log)
	loginSvc := appHandler.New(apiClient, log)
	slotsSvc := appSlots.New(apiClient, log)
	recordsSvc := appRecords.New(apiClient, log)

	serverBot := botServer.NewServer(b, log, loginSvc, slotsSvc, recordsSvc, "bot")
	log.Info("cmd.launchBot: registering bot command handlers")
	serverBot.RegisterHandlers()

	botHandler := msgHandler.NewHandler(b, log)

	httpServer := http.Server{
		Addr:    ":8091",
		Handler: http.NewServeMux(),
	}
	httpClient := httpTransport.NewHttpClient("server", &httpServer, b, botHandler, log)

	manager := closer.NewManager(log)
	manager.AddGraceful(httpClient)
	manager.AddGraceful(closeRateLimiter)
	manager.AddCloser(stateManager)
	manager.AddGraceful(serverBot)
	go func() {
		if err := httpClient.Run(); err != nil {
			log.Infof("http-client error: %v", err)
			cancel()
		}

		log.Info("cmd.launchBot: HTTP notifier server started on :8091")

	}()
	go func() {
		if err := serverBot.Start(ctx); err != nil {
			log.Infof("telegram bot error: %v", err)
			cancel()
		}
		log.Infof("cmd.launchBot: starting bot polling")
	}()

	time.Sleep(2 * time.Second)
	manager.WaitForSignal()
	log.Infof("cmd.launchBot: Stopped")
}
