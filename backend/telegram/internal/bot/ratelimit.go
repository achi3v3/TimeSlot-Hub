package bot

import (
	"context"
	"fmt"
	"sync"
	"telegram-bot/internal/adapter/backendapi"
	"telegram-bot/internal/handlers/components"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// RateLimiter ограничивает частоту выполнения команд и callback'ов
type RateLimiter struct {
	mu       sync.RWMutex
	users    map[int64]*UserLimits
	cleanup  *time.Ticker
	stopChan chan bool
}

// UserLimits хранит ограничения для конкретного пользователя
type UserLimits struct {
	LastCommand   time.Time
	LastCallback  time.Time
	CommandCount  int
	CallbackCount int
	LastReset     time.Time
}

// RateLimitConfig конфигурация ограничений
type RateLimitConfig struct {
	CommandCooldown       time.Duration // Минимальный интервал между командами
	CallbackCooldown      time.Duration // Минимальный интервал между callback'ами
	MaxCommandsPerMinute  int           // Максимум команд в минуту
	MaxCallbacksPerMinute int           // Максимум callback'ов в минуту
	CleanupInterval       time.Duration // Интервал очистки старых записей
}

// DefaultRateLimitConfig возвращает конфигурацию по умолчанию
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		CommandCooldown:       1500 * time.Millisecond, // 2 секунды между командами
		CallbackCooldown:      500 * time.Millisecond,  // 1 секунда между callback'ами
		MaxCommandsPerMinute:  10,                      // 10 команд в минуту
		MaxCallbacksPerMinute: 30,                      // 30 callback'ов в минуту
		CleanupInterval:       5 * time.Minute,         // Очистка каждые 5 минут
	}
}

var (
	rateLimiter *RateLimiter
	config      = DefaultRateLimitConfig()
)

// InitRateLimiter инициализирует глобальный rate limiter
func InitRateLimiter() {
	rateLimiter = &RateLimiter{
		users:    make(map[int64]*UserLimits),
		cleanup:  time.NewTicker(config.CleanupInterval),
		stopChan: make(chan bool),
	}

	// Запускаем горутину для очистки старых записей
	go rateLimiter.cleanupRoutine()
}

type CloseRateLimiter struct {
	name string
}

func NewCloseRateLimiter(name string) *CloseRateLimiter {
	return &CloseRateLimiter{
		name: name,
	}
}
func (crl *CloseRateLimiter) Name() string { return crl.name }
func (crl *CloseRateLimiter) Shutdown(ctx context.Context) error {
	StopRateLimiter()
	return nil
}

// StopRateLimiter останавливает rate limiter
func StopRateLimiter() {
	if rateLimiter != nil {
		rateLimiter.Shutdown()
	}
}

// cleanupRoutine периодически очищает старые записи пользователей
func (rl *RateLimiter) cleanupRoutine() {
	for {
		select {
		case <-rl.cleanup.C:
			rl.cleanupOldUsers()
		case <-rl.stopChan:
			rl.cleanup.Stop()
			return
		}
	}
}

// cleanupOldUsers удаляет записи пользователей, которые не активны более часа
func (rl *RateLimiter) cleanupOldUsers() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-time.Hour)
	for userID, limits := range rl.users {
		if limits.LastReset.Before(cutoff) {
			delete(rl.users, userID)
		}
	}
}

// stop останавливает rate limiter
func (rl *RateLimiter) Shutdown() {
	if rl == nil {
		return
	}
	rl.stopChan <- true

	rl.mu.Lock()
	rl.users = nil
	rl.mu.Unlock()
}

// CanExecuteCommand проверяет, может ли пользователь выполнить команду
func CanExecuteCommand(userID int64) bool {
	if rateLimiter == nil {
		return true // Если rate limiter не инициализирован, разрешаем
	}
	return rateLimiter.canExecuteCommand(userID)
}

func (rl *RateLimiter) canExecuteCommand(userID int64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limits, exists := rl.users[userID]

	if !exists {
		limits = &UserLimits{
			LastCommand:  now,
			LastReset:    now,
			CommandCount: 1,
		}
		rl.users[userID] = limits
		return true
	}

	// Проверяем cooldown между командами
	if now.Sub(limits.LastCommand) < config.CommandCooldown {
		return false
	}

	// Сбрасываем счетчик если прошла минута
	if now.Sub(limits.LastReset) >= time.Minute {
		limits.CommandCount = 0
		limits.LastReset = now
	}

	// Проверяем лимит команд в минуту
	if limits.CommandCount >= config.MaxCommandsPerMinute {
		return false
	}

	limits.LastCommand = now
	limits.CommandCount++
	return true
}

// CanExecuteCallback проверяет, может ли пользователь выполнить callback
func CanExecuteCallback(userID int64) bool {
	if rateLimiter == nil {
		return true // Если rate limiter не инициализирован, разрешаем
	}
	return rateLimiter.canExecuteCallback(userID)
}

func (rl *RateLimiter) canExecuteCallback(userID int64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limits, exists := rl.users[userID]

	if !exists {
		limits = &UserLimits{
			LastCallback:  now,
			LastReset:     now,
			CallbackCount: 1,
		}
		rl.users[userID] = limits
		return true
	}

	// Проверяем cooldown между callback'ами
	if now.Sub(limits.LastCallback) < config.CallbackCooldown {
		return false
	}

	// Сбрасываем счетчик если прошла минута
	if now.Sub(limits.LastReset) >= time.Minute {
		limits.CallbackCount = 0
		limits.LastReset = now
	}

	// Проверяем лимит callback'ов в минуту
	if limits.CallbackCount >= config.MaxCallbacksPerMinute {
		return false
	}

	limits.LastCallback = now
	limits.CallbackCount++
	return true
}

// CommandRateLimitMiddleware middleware для ограничения команд
func CommandRateLimitMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update == nil || update.Message == nil || update.Message.From == nil {
			return
		}

		userID := update.Message.From.ID

		if !CanExecuteCommand(userID) {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: userID,
				Text:   "⏳ Слишком быстро! Подождите немного перед следующей командой.",
			})
			return
		}

		next(ctx, b, update)
	}
}

// CallbackRateLimitMiddleware middleware для ограничения callback'ов
func CallbackRateLimitMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update == nil || update.CallbackQuery == nil {
			return
		}

		userID := update.CallbackQuery.From.ID

		if !CanExecuteCallback(userID) {
			b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            "⏳ Слишком быстро! Подождите немного.",
				ShowAlert:       true,
			})
			return
		}

		next(ctx, b, update)
	}
}

// CommandAuthMiddleware middleware для ограничения команд для не зарегистрированных пользователей
func CommandAuthMiddleware(next bot.HandlerFunc, client *backendapi.Client) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update == nil || update.Message == nil || update.Message.From == nil {
			return
		}

		userID := update.Message.From.ID

		_, exist := client.CheckAuth(ctx, userID)
		if !exist {
			msgText := fmt.Sprintf(
				"%s"+
					"<blockquote> ℹ️ Для начала использования бота, вам необходимо зарегистрироваться!\n\nДля регистрации /start</blockquote>",
				components.Header(),
			)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    userID,
				ParseMode: models.ParseModeHTML,
				Text:      msgText,
			})
			return
		}

		next(ctx, b, update)
	}
}
