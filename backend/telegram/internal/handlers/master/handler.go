package master

import (
	"context"
	"fmt"
	"strings"
	adapter "telegram-bot/internal/adapter/backendapi"
	"telegram-bot/internal/config"
	"telegram-bot/internal/handlers/components"
	"telegram-bot/internal/utils"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	logger *logrus.Logger
	client *adapter.Client
}

func NewHandler(logger *logrus.Logger) *Handler {
	cfg := config.Load()
	client := adapter.New(cfg.BackendBaseURL, logger)
	return &Handler{logger: logger, client: client}
}

// HandlerUpcomingRecords показывает предстоящие подтвержденные записи мастера
func (h *Handler) HandlerUpcomingRecords(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	telegramID := update.Message.From.ID

	h.logger.WithField("telegram_id", telegramID).Info("Handler.UpcomingRecords: fetching upcoming records")

	// Получаем предстоящие записи из backend
	records, err := h.client.GetUpcomingRecordsByMasterTelegramID(ctx, telegramID)
	if err != nil {
		h.logger.Errorf("Handler.UpcomingRecords: failed to get records: %v", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      fmt.Sprintf("%s❌ Не удалось получить список записей", components.Header()),
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	if len(records) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      fmt.Sprintf("%s<b>Предстоящие записи</b>\n\n<i>У вас пока нет предстоящих подтвержденных записей</i>", components.Header()),
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	// Показываем первую страницу (страница 1)
	h.showUpcomingRecordsPage(ctx, b, chatID, telegramID, records, 1, 0)
}

// showUpcomingRecordsPage показывает страницу записей с пагинацией
func (h *Handler) showUpcomingRecordsPage(ctx context.Context, b *bot.Bot, chatID int64, telegramID int64, records []map[string]interface{}, page int, messageID int) {
	const limit = 5

	// Форматируем список записей для текущей страницы
	text, totalPages := h.FormatUpcomingRecordsPage(records, page, limit, telegramID)

	// Создаем клавиатуру для пагинации
	keyboard := h.BuildUpcomingRecordsPagination(page, totalPages, telegramID)

	if messageID > 0 {
		// Редактируем существующее сообщение
		_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   messageID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: keyboard,
		})
		if err != nil {
			h.logger.Errorf("Handler.UpcomingRecords: failed to edit message: %v", err)
		}
	} else {
		// Отправляем новое сообщение
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: keyboard,
		})
	}
}

// FormatUpcomingRecordsPage форматирует страницу записей с пагинацией (экспортируемая)
func (h *Handler) FormatUpcomingRecordsPage(records []map[string]interface{}, page int, limit int, telegramID int64) (string, int) {
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("%s<b>📅 Предстоящие записи</b>\n\n", components.Header()))

	if len(records) == 0 {
		b.WriteString("<i>У вас нет предстоящих записей</i>")
		return b.String(), 1
	}

	// Пагинация: выбираем записи для текущей страницы
	start := (page - 1) * limit
	end := start + limit
	if end > len(records) {
		end = len(records)
	}

	totalPages := (len(records) + limit - 1) / limit
	pageRecords := records[start:end]

	// Группируем записи по датам
	type dateGroup struct {
		date    string
		records []map[string]interface{}
	}

	dateGroups := make(map[string]*dateGroup)
	dateOrder := []string{}

	for _, r := range pageRecords {
		slotData, _ := r["slot"].(map[string]interface{})
		masterData, _ := slotData["master"].(map[string]interface{})

		// Получаем таймзону мастера
		masterTimezone := ""
		if tz, ok := masterData["timezone"].(string); ok {
			masterTimezone = tz
		}
		if masterTimezone == "" {
			masterTimezone = "Europe/Moscow"
		}

		// Получаем дату
		startTimeStr, _ := slotData["start_time"].(string)
		date := "не задано"
		if startTimeStr != "" {
			if startTime, err := parseTime(startTimeStr); err == nil && !startTime.IsZero() {
				date = utils.FormatDateInLocation(masterTimezone, startTime)
			}
		}

		// Добавляем запись в группу по дате
		if _, exists := dateGroups[date]; !exists {
			dateGroups[date] = &dateGroup{date: date, records: []map[string]interface{}{}}
			dateOrder = append(dateOrder, date)
		}
		dateGroups[date].records = append(dateGroups[date].records, r)
	}

	// Выводим записи по датам
	for i, date := range dateOrder {
		group := dateGroups[date]

		// Заголовок даты
		b.WriteString(fmt.Sprintf("<b>📆 %s</b>\n\n", group.date))

		// Записи в этой дате
		for j, r := range group.records {
			slotData, _ := r["slot"].(map[string]interface{})
			clientData, _ := r["client"].(map[string]interface{})
			serviceData, _ := slotData["service"].(map[string]interface{})
			masterData, _ := slotData["master"].(map[string]interface{})

			// Получаем таймзону мастера
			masterTimezone := ""
			if tz, ok := masterData["timezone"].(string); ok {
				masterTimezone = tz
			}
			if masterTimezone == "" {
				masterTimezone = "Europe/Moscow"
			}

			// Форматируем времена
			startTimeStr, _ := slotData["start_time"].(string)
			endTimeStr, _ := slotData["end_time"].(string)

			start := "--:--"
			end := "--:--"

			if startTimeStr != "" {
				if startTime, err := parseTime(startTimeStr); err == nil && !startTime.IsZero() {
					start = utils.FormatTimeOnlyInLocation(masterTimezone, startTime)
				}
			}

			if endTimeStr != "" {
				if endTime, err := parseTime(endTimeStr); err == nil && !endTime.IsZero() {
					end = utils.FormatTimeOnlyInLocation(masterTimezone, endTime)
				}
			}

			// Получаем смещение таймзоны
			tzOffset := utils.GetTimezoneOffset(masterTimezone)

			// Получаем имя клиента
			clientName := "Неизвестный клиент"
			if firstName, ok := clientData["first_name"].(string); ok && firstName != "" {
				if surname, ok := clientData["surname"].(string); ok && surname != "" {
					clientName = fmt.Sprintf("%s %s", firstName, surname)
				} else {
					clientName = firstName
				}
			}

			// Получаем телефон клиента
			clientPhone := ""
			if phone, ok := clientData["phone"].(string); ok {
				clientPhone = phone
			}

			// Получаем название услуги
			serviceName := "Неизвестная услуга"
			if name, ok := serviceData["name"].(string); ok && name != "" {
				serviceName = name
			}

			// Получаем длительность и цену услуги
			duration := 0
			if d, ok := serviceData["duration"].(float64); ok {
				duration = int(d)
			}

			// Формируем текст записи
			b.WriteString("<blockquote>")
			b.WriteString(fmt.Sprintf("<b>🕐 Время:</b> <code>%s - %s (TZ: %s %s)</code>\n", start, end, masterTimezone, tzOffset))
			b.WriteString(fmt.Sprintf("<b>📍 Услуга:</b> <code>%s</code>\n", serviceName))
			b.WriteString(fmt.Sprintf("<b>🧍🏼 Клиент:</b> <code>%s</code> (<code>%s</code>)\n", clientName, clientPhone))
			b.WriteString(fmt.Sprintf("<b>⏱ Длительность:</b> <code>%d мин.</code>\n", duration))
			b.WriteString("</blockquote>")

			// Добавляем разделитель между записями одного дня
			if j < len(group.records)-1 {
				b.WriteString("\n")
			}
		}

		// Добавляем пустую строку между разными датами
		if i < len(dateOrder)-1 {
			b.WriteString("\n\n")
		}
	}

	// Добавляем информацию о странице
	b.WriteString(fmt.Sprintf("\n\n<i>Страница %d из %d</i>", page, totalPages))

	return b.String(), totalPages
}

// BuildUpcomingRecordsPagination создает клавиатуру для пагинации (экспортируемая)
func (h *Handler) BuildUpcomingRecordsPagination(page int, totalPages int, telegramID int64) *models.InlineKeyboardMarkup {
	buttons := [][]models.InlineKeyboardButton{}

	row := []models.InlineKeyboardButton{}

	// Кнопка "Назад"
	if page > 1 {
		row = append(row, models.InlineKeyboardButton{
			Text:         "⬅️ Назад",
			CallbackData: fmt.Sprintf("upcoming_page/%d/%d", page-1, telegramID),
		})
	}

	// Кнопка "Вперед"
	if page < totalPages {
		row = append(row, models.InlineKeyboardButton{
			Text:         "Вперед ➡️",
			CallbackData: fmt.Sprintf("upcoming_page/%d/%d", page+1, telegramID),
		})
	}

	if len(row) > 0 {
		buttons = append(buttons, row)
	}

	return &models.InlineKeyboardMarkup{InlineKeyboard: buttons}
}

// parseTime парсит строку времени в формате ISO 8601
func parseTime(s string) (time.Time, error) {
	// Пробуем различные форматы
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}
