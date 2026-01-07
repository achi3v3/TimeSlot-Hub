package utils

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Конфигурируемая временная зона для форматирования сообщений в Telegram.
// Используем TELEGRAM_TIMEZONE или TIMEZONE, иначе системную time.Local.
var (
	tzOnce sync.Once
	tzLoc  *time.Location
)

func getTZ() *time.Location {
	tzOnce.Do(func() {
		name := os.Getenv("TELEGRAM_TIMEZONE")
		if name == "" {
			name = os.Getenv("TIMEZONE")
		}
		if name != "" {
			if loc, err := time.LoadLocation(name); err == nil {
				tzLoc = loc
			}
		}
		if tzLoc == nil {
			tzLoc = time.Local
		}
	})
	return tzLoc
}

func toTZ(t time.Time) time.Time { return t.In(getTZ()) }

// Исторические имена функций оставляем, но теперь форматируем в выбранной TZ
func FormatTimeInMoscow(t time.Time, layout string) string { return toTZ(t).Format(layout) }
func FormatDateInMoscow(t time.Time) string                { return FormatTimeInMoscow(t, "02-01-2006") }
func FormatTimeOnlyInMoscow(t time.Time) string            { return FormatTimeInMoscow(t, "15:04") }

// Новые функции: форматирование по имени таймзоны (например, "Europe/Moscow")
func FormatDateInLocation(locName string, t time.Time) string {
	loc := loadLocationOrDefault(locName)
	return t.In(loc).Format("02-01-2006")
}

func FormatTimeOnlyInLocation(locName string, t time.Time) string {
	loc := loadLocationOrDefault(locName)
	return t.In(loc).Format("15:04")
}

func loadLocationOrDefault(locName string) *time.Location {
	if locName == "" {
		// Если таймзона не указана, используем московскую
		if moscowLoc, err := time.LoadLocation("Europe/Moscow"); err == nil {
			return moscowLoc
		}
		return getTZ()
	}
	if loc, err := time.LoadLocation(locName); err == nil {
		return loc
	}
	// Если указанная таймзона не загрузилась, используем московскую
	if moscowLoc, err := time.LoadLocation("Europe/Moscow"); err == nil {
		return moscowLoc
	}
	return getTZ()
}

// GetTimezoneOffset возвращает смещение в часах для указанной таймзоны в формате (+3) или (+5)
func GetTimezoneOffset(locName string) string {
	loc := loadLocationOrDefault(locName)
	now := time.Now().In(loc)
	_, offset := now.Zone()
	hours := offset / 3600
	if hours >= 0 {
		return fmt.Sprintf("+%d", hours)
	}
	return fmt.Sprintf("%d", hours)
}
