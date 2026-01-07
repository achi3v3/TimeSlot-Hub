package main

import (
	"app/http/router"
	"app/pkg/closer"
	"context"
	"net/http"
	"time"

	"app/internal/database"
	reminder "app/internal/scheduler"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	database.Init()
	db := database.GetDB()

	r := gin.Default()
	httpServer := router.NewHTTPServer(&http.Server{
		Addr:    ":8090",
		Handler: r,
	}, "server")

	client := router.NewClient(db.DB, logger, httpServer, r)

	// Start background reminders (1-hour before confirmed records)
	reminderCtx, stopReminder := context.WithCancel(ctx)
	defer stopReminder()
	rem := reminder.NewReminder(db.DB, logger)
	rem.StartReminder(reminderCtx)

	manager := closer.NewManager(logger)
	manager.AddGraceful(httpServer)
	manager.AddGraceful(reminder.NewReminderCloser(stopReminder, "reminder-scheduler"))
	manager.AddGraceful(db)

	go func() {
		if err := client.Run(); err != nil {
			logger.Errorf("Server failed: %v", err)
			cancel()
		}
	}()

	time.Sleep(2 * time.Second)

	manager.WaitForSignal()

}
