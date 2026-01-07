package reminder

import (
	"app/http/repository/notification"
	recrepo "app/http/repository/record"
	"app/http/sender"
	"app/pkg/models"
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Reminder struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewReminder(db *gorm.DB, logger *logrus.Logger) *Reminder {
	return &Reminder{
		db:     db,
		logger: logger,
	}
}

// StartReminder launches a lightweight ticker that sends 1-hour reminders for confirmed records.
func (r *Reminder) StartReminder(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		recordRepo := recrepo.NewRepository(r.db, r.logger)
		notifRepo := notification.NewRepository(r.db, r.logger)
		for {
			select {
			case <-ticker.C:
				r.sendReminders(recordRepo, notifRepo, r.logger)
			case <-ctx.Done():
				r.logger.Info("Reminder stopped")
				return
			}
		}
	}()
}

func (r *Reminder) sendReminders(recordRepo *recrepo.Repository, notifRepo *notification.Repository, logger *logrus.Logger) {
	// Narrow 2-minute window around now+60m to avoid duplicates
	now := time.Now().UTC()
	start := now.Add(59 * time.Minute)
	end := now.Add(61 * time.Minute)

	records, err := recordRepo.FindConfirmedRecordsStartingBetween(start, end)
	if err != nil {
		logger.WithError(err).Warn("reminder: query failed")
		return
	}
	if len(records) == 0 {
		return
	}
	for _, r := range records {
		clientTg := r.Client.TelegramID
		if clientTg == 0 {
			continue
		}
		tz := r.Slot.Master.Timezone
		if tz == "" {
			tz = "Europe/Moscow"
		}
		loc, err := time.LoadLocation(tz)
		if err != nil {
			loc = time.FixedZone("Europe/Moscow", 3*3600)
		}
		startAt := r.Slot.StartTime.In(loc)
		endAt := r.Slot.EndTime
		if endAt.IsZero() && r.Slot.Service.Duration > 0 {
			endAt = r.Slot.StartTime.Add(time.Duration(r.Slot.Service.Duration) * time.Minute)
		}
		endAt = endAt.In(loc)

		dateText := startAt.Format("02.01.2006")
		timeText := startAt.Format("15:04") + " - " + endAt.Format("15:04") + " (TZ: " + tz + ")"

		masterName := r.Slot.Master.FirstName + " " + r.Slot.Master.Surname
		serviceName := r.Slot.Service.Name

		title := "Напоминание: запись через 1 час"
		message := "У вас запись к: " + masterName + "\n" +
			"Услуга: " + serviceName + "\n" +
			"Дата: " + dateText + "\n" +
			"Время: " + timeText

		if err := sender.RecordStatusNotify(clientTg, title, message); err != nil {
			logger.WithError(err).Warn("reminder: telegram notify failed")
			continue
		}

		// Also create an in-app notification for the client
		notif := &models.Notification{
			UserID:  r.Client.ID,
			Type:    "RECORD_REMINDER_1H",
			Title:   title,
			Message: message,
		}
		if err := notifRepo.Create(notif); err != nil {
			logger.WithError(err).Warn("reminder: create frontend notification failed")
		}
	}
}

type ReminderCloser struct {
	stop context.CancelFunc
	name string
}

func NewReminderCloser(stop context.CancelFunc, name string) *ReminderCloser {
	return &ReminderCloser{
		stop: stop,
		name: name,
	}
}

func (r *ReminderCloser) Shutdown(ctx context.Context) error {
	r.stop()

	// waiting
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		// stopped
	}
	return nil
}

func (r *ReminderCloser) Name() string {
	return r.name
}
