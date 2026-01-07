package record

import (
	"context"
	"fmt"
	"sort"
	"strings"
	adapter "telegram-bot/internal/adapter/backendapi"
	"telegram-bot/internal/config"
	"telegram-bot/internal/handlers/components"
	"telegram-bot/internal/handlers/message"
	"telegram-bot/internal/utils"
	mymodels "telegram-bot/pkg/models"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"
)

type Service struct {
	bot    *bot.Bot
	logger *logrus.Logger
	client *adapter.Client
	editor *message.MessageEditor
}

func NewService(bot *bot.Bot, logger *logrus.Logger) *Service {
	cfg := config.Load()
	client := adapter.New(cfg.BackendBaseURL, logger)
	return &Service{bot: bot, logger: logger, client: client, editor: message.NewMessageEditor()}
}

func (s *Service) SendUserRecords(ctx context.Context, chatID int64, telegramID int64, status string, page int) {
	if page <= 0 {
		page = 1
	}
	const limit = 10

	// Определяем статус для запроса
	queryStatus := status
	if status == "all" {
		queryStatus = ""
	}

	resp, err := s.client.GetUserRecordsFiltered(ctx, telegramID, queryStatus, 1, 1000)
	if err != nil {
		s.logger.WithError(err).Warn("Handlers.Record.SendUserRecords: filter request failed")
		s.bot.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "❌ Не удалось получить записи"})
		return
	}

	// Фильтруем только будущие записи
	filteredRecords := s.filterRecordsByStatus(resp.Records, queryStatus)
	future, _ := splitRecordsByTime(filteredRecords)

	if len(future) == 0 {
		statusText := "записи"
		if status != "" && status != "all" {
			switch status {
			case "confirm":
				statusText = "подтвержденные записи"
			case "reject":
				statusText = "отклоненные записи"
			case "pending":
				statusText = "записи в ожидании"
			}
		}
		text := fmt.Sprintf("%s<b>Мои записи (%s)</b>\n<i>У вас пока нет предстоящих %s</i>", components.Header(), statusText, statusText)
		s.bot.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: text, ParseMode: models.ParseModeHTML})
		return
	}

	// Показываем только предстоящие записи с пагинацией
	pages := pagesCount(len(future), limit)
	start := (page - 1) * limit
	if start > len(future) {
		start = len(future)
	}
	end := start + limit
	if end > len(future) {
		end = len(future)
	}
	pageRecords := future[start:end]

	statusText := "записи"
	if status != "" && status != "all" {
		switch status {
		case "confirm":
			statusText = "подтвержденные записи"
		case "reject":
			statusText = "отклоненные записи"
		case "pending":
			statusText = "записи в ожидании"
		}
	}

	text := fmt.Sprintf("%s<b>Мои записи (%s)</b> (стр. %d/%d)\n\n", components.Header(), statusText, page, pages) + s.formatRecordsText(pageRecords)
	keyboard := s.buildRecordsPaginationKeyboard(status, page, pages)

	sent, err := s.bot.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: text, ParseMode: models.ParseModeHTML, ReplyMarkup: keyboard})
	if err == nil {
		sm := message.GetStateManager()
		sm.RemoveMessageState(chatID, chatID, "user_records")
		sm.SetMessageState(chatID, chatID, sent.ID, "user_records", map[string]interface{}{"timeType": "future_only", "page": page, "status": status})
	}
}

// SendAllRecords отправляет все записи с меню выбора будущих/прошедших
func (s *Service) SendAllRecords(ctx context.Context, chatID int64, telegramID int64) {
	const limit = 10
	_, err := s.client.GetUserRecordsFiltered(ctx, telegramID, "", 1, 1000)
	if err != nil {
		s.logger.WithError(err).Warn("Handlers.Record.SendAllRecords: filter request failed")
		s.bot.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "❌ Не удалось получить записи"})
		return
	}

	// Показываем меню выбора времени (Будущие/Прошедшие)
	text := fmt.Sprintf("%s<b>Все мои записи</b>\n<i>Выберите пул записей для просмотра</i>", components.Header())
	keyboard := &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
		{{Text: "Будущие", CallbackData: "all_records_time/future/all/1"}},
		{{Text: "Прошедшие", CallbackData: "all_records_time/past/all/1"}},
	}}

	// Всегда новое сообщение и установка указателя состояния
	sent, err := s.bot.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: text, ParseMode: models.ParseModeHTML, ReplyMarkup: keyboard})
	if err == nil {
		sm := message.GetStateManager()
		sm.RemoveMessageState(chatID, chatID, "all_records")
		sm.SetMessageState(chatID, chatID, sent.ID, "all_records", map[string]interface{}{"timeType": "chooser", "page": 1, "status": "all"})
	}
}

// EditUserRecordsPage редактирует текущее сообщение с записями (используется при нажатии на кнопки пагинации)
func (s *Service) EditUserRecordsPage(ctx context.Context, chatID int64, telegramID int64, status string, page int, messageID int) {
	if page <= 0 {
		page = 1
	}
	const limit = 10

	// Определяем статус для запроса
	queryStatus := status
	if status == "all" {
		queryStatus = ""
	}

	resp, err := s.client.GetUserRecordsFiltered(ctx, telegramID, queryStatus, 1, 1000)
	if err != nil {
		s.logger.WithError(err).Warn("Handlers.Record.EditUserRecordsPage: filter request failed")
		return
	}

	// Фильтруем только будущие записи
	filteredRecords := s.filterRecordsByStatus(resp.Records, queryStatus)
	future, _ := splitRecordsByTime(filteredRecords)

	pages := pagesCount(len(future), limit)
	start := (page - 1) * limit
	if start > len(future) {
		start = len(future)
	}
	end := start + limit
	if end > len(future) {
		end = len(future)
	}
	pageRecords := future[start:end]

	statusText := "записи"
	if status != "" && status != "all" {
		switch status {
		case "confirm":
			statusText = "подтвержденные записи"
		case "reject":
			statusText = "отклоненные записи"
		case "pending":
			statusText = "записи в ожидании"
		}
	}

	text := fmt.Sprintf("%s<b>Мои записи (%s)</b> (стр. %d/%d)\n\n", components.Header(), statusText, page, pages) + s.formatRecordsText(pageRecords)
	keyboard := s.buildRecordsPaginationKeyboard(status, page, pages)

	// Редактируем конкретное сообщение, если messageID > 0, иначе отправляем новое
	if messageID > 0 {
		err := s.editor.EditSpecificMessage(ctx, s.bot, chatID, messageID, text, keyboard)
		if err != nil {
			s.logger.WithError(err).Error("Failed to edit records page message")
		}
	} else {
		// Отправляем новое сообщение
		sent, err := s.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: keyboard,
		})
		if err != nil {
			s.logger.WithError(err).Error("Failed to send records page message")
			return
		}

		// Устанавливаем состояние для нового сообщения
		sm := message.GetStateManager()
		sm.RemoveMessageState(chatID, chatID, "user_records")
		sm.SetMessageState(chatID, chatID, sent.ID, "user_records", map[string]interface{}{"timeType": "future_only", "page": page, "status": status})
	}
}

// EditUserRecordsTimePage редактирует текущее сообщение: будущее/прошедшее, с сортировкой и пагинацией
func (s *Service) EditUserRecordsTimePage(ctx context.Context, chatID int64, telegramID int64, timeType string, status string, page int, messageID int) {
	if page <= 0 {
		page = 1
	}
	const limit = 10

	// Определяем статус для запроса
	queryStatus := status
	if status == "all" {
		queryStatus = ""
	}

	resp, err := s.client.GetUserRecordsFiltered(ctx, telegramID, queryStatus, 1, 1000)
	if err != nil {
		s.logger.WithError(err).Warn("Handlers.Record.EditUserRecordsTimePage: load failed")
		return
	}

	// Сначала фильтруем по статусу, затем по времени
	filteredRecords := s.filterRecordsByStatus(resp.Records, queryStatus)
	future, past := splitRecordsByTime(filteredRecords)

	var list []mymodels.Record
	var title string
	if timeType == "past" {
		list = past
		title = "Прошедшие записи"
	} else {
		list = future
		title = "Будущие записи"
	}

	// Добавляем информацию о статусе в заголовок
	if status != "" && status != "all" {
		statusText := ""
		switch status {
		case "confirm":
			statusText = " (подтвержденные)"
		case "reject":
			statusText = " (отклоненные)"
		case "pending":
			statusText = " (в ожидании)"
		}
		title += statusText
	}

	pages := pagesCount(len(list), limit)
	start := (page - 1) * limit
	if start > len(list) {
		start = len(list)
	}
	end := start + limit
	if end > len(list) {
		end = len(list)
	}
	pageRecords := list[start:end]
	text := fmt.Sprintf("%s<b>%s</b> (стр. %d/%d)\n\n", components.Header(), title, page, pages) + s.formatRecordsText(pageRecords)
	keyboard := s.buildRecordsTimePaginationKeyboard(timeType, status, page, pages)

	// Редактируем конкретное сообщение, если messageID > 0, иначе отправляем новое
	if messageID > 0 {
		err := s.editor.EditSpecificMessage(ctx, s.bot, chatID, messageID, text, keyboard)
		if err != nil {
			s.logger.WithError(err).Error("Failed to edit records time page message")
		}
	} else {
		// Отправляем новое сообщение
		sent, err := s.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: keyboard,
		})
		if err != nil {
			s.logger.WithError(err).Error("Failed to send records time page message")
			return
		}

		// Устанавливаем состояние для нового сообщения
		sm := message.GetStateManager()
		sm.RemoveMessageState(chatID, chatID, "user_records")
		sm.SetMessageState(chatID, chatID, sent.ID, "user_records", map[string]interface{}{"timeType": timeType, "page": page, "status": status})
	}
}

func (s *Service) formatRecordsText(records []mymodels.Record) string {
	b := strings.Builder{}
	// b.WriteString(fmt.Sprintf("%sВаши записи (стр. %d/%d)\n\n", components.Header(), page, total))

	if len(records) == 0 {
		b.WriteString("<i>У вас нет записей.</i>")
		return b.String()
	}

	for _, r := range records {
		// Получаем таймзону мастера из Slot.Master.Timezone
		masterTimezone := r.Slot.Master.Timezone
		// Если таймзона не указана, используем московскую как fallback
		if masterTimezone == "" {
			masterTimezone = "Europe/Moscow"
		}

		// Защита от нулевых дат: если время по нулям — выводим "не задано"
		date := "не задано"
		start := "--:--"
		end := "--:--"
		if !r.Slot.StartTime.IsZero() {
			// Используем таймзону мастера для форматирования
			date = utils.FormatDateInLocation(masterTimezone, r.Slot.StartTime)
			start = utils.FormatTimeOnlyInLocation(masterTimezone, r.Slot.StartTime)
		}
		if !r.Slot.EndTime.IsZero() {
			end = utils.FormatTimeOnlyInLocation(masterTimezone, r.Slot.EndTime)
		}

		// Определяем статус записи с эмодзи
		statusEmoji := "⏳"
		statusText := "Ожидает подтверждения"
		switch r.Status {
		case "confirm":
			statusEmoji = "✅"
			statusText = "Подтверждена"
		case "reject":
			statusEmoji = "❌"
			statusText = "Отклонена"
		case "pending":
			statusEmoji = "⏳"
			statusText = "В ожидании"
		}

		// Получаем информацию о мастере и услуге
		masterName := "Неизвестный мастер"
		serviceName := "Неизвестная услуга"

		if r.Slot.Master.FirstName != "" || r.Slot.Master.Surname != "" {
			masterName = fmt.Sprintf("%s %s", r.Slot.Master.FirstName, r.Slot.Master.Surname)
		}

		if r.Slot.Service.Name != "" {
			serviceName = r.Slot.Service.Name
		}

		// Получаем смещение таймзоны для отображения
		tzOffset := utils.GetTimezoneOffset(masterTimezone)

		// Формируем текст времени с указанием таймзоны (всегда показываем таймзону)
		timeText := fmt.Sprintf("%s - %s (TZ: %s %s)", start, end, masterTimezone, tzOffset)

		// Формируем текст даты с указанием таймзоны
		dateText := date
		if date != "не задано" {
			dateText = fmt.Sprintf("%s (TZ: %s %s)", date, masterTimezone, tzOffset)
		}

		b.WriteString(fmt.Sprintf(
			"<blockquote>"+
				"<code>%s — %s</code>\n"+
				// "<b>Мастер:</b> <code>%s</code>\n"+
				"<b>Статус:</b> <code>%s %s</code>\n"+
				"<b>Дата:</b> <code>%s</code>\n"+
				"<b>Время:</b> <code>%s</code>\n"+
				"</blockquote>",
			serviceName,
			masterName,
			statusEmoji,
			statusText,
			dateText,
			timeText,
		))
	}
	return b.String()
}

// buildRecordsPaginationKeyboard создает inline-кнопки для листания страниц записей
func (s *Service) buildRecordsPaginationKeyboard(status string, page, total int) *models.InlineKeyboardMarkup {
	buttons := [][]models.InlineKeyboardButton{}
	// prev
	if page > 1 {
		buttons = append(buttons, []models.InlineKeyboardButton{{
			Text:         "⬅️ Предыдущая",
			CallbackData: fmt.Sprintf("records/%s/%d", safeStatus(status), page-1),
		}})
	}
	// next
	if page < total {
		buttons = append(buttons, []models.InlineKeyboardButton{{
			Text:         "Следующая ➡️",
			CallbackData: fmt.Sprintf("records/%s/%d", safeStatus(status), page+1),
		}})
	}
	return &models.InlineKeyboardMarkup{InlineKeyboard: buttons}
}

func safeStatus(status string) string {
	if status == "" {
		return "all"
	}
	return status
}
func (s *Service) buildRecordsTimePaginationKeyboard(timeType string, status string, page, total int) *models.InlineKeyboardMarkup {
	buttons := [][]models.InlineKeyboardButton{}
	if page > 1 {
		buttons = append(buttons, []models.InlineKeyboardButton{{
			Text: "⬅️ Предыдущая", CallbackData: fmt.Sprintf("records_time/%s/%s/%d", timeType, status, page-1),
		}})
	}
	if page < total {
		buttons = append(buttons, []models.InlineKeyboardButton{{
			Text: "Следующая ➡️", CallbackData: fmt.Sprintf("records_time/%s/%s/%d", timeType, status, page+1),
		}})
	}
	// Добавим кнопку вернуться к выбору пула
	buttons = append(buttons, []models.InlineKeyboardButton{{Text: "◀️ К выбору", CallbackData: "records_time/chooser/1"}})
	return &models.InlineKeyboardMarkup{InlineKeyboard: buttons}
}

func splitRecordsByTime(records []mymodels.Record) (future []mymodels.Record, past []mymodels.Record) {
	now := time.Now()
	threshold := time.Hour // 1 час просрочки
	for _, r := range records {
		start := r.Slot.StartTime
		end := r.Slot.EndTime
		if end.IsZero() && !start.IsZero() && r.Slot.Service.Duration > 0 {
			end = start.Add(time.Duration(r.Slot.Service.Duration) * time.Minute)
		}
		isPast := false
		if !end.IsZero() {
			isPast = end.Add(threshold).Before(now)
		} else if !start.IsZero() {
			isPast = start.Add(threshold).Before(now)
		}
		if isPast {
			past = append(past, r)
		} else {
			future = append(future, r)
		}
	}
	// Сортировка: будущие по возрастанию start, прошедшие по убыванию end
	sort.Slice(future, func(i, j int) bool { return future[i].Slot.StartTime.Before(future[j].Slot.StartTime) })
	sort.Slice(past, func(i, j int) bool {
		ei := past[i].Slot.EndTime
		ej := past[j].Slot.EndTime
		if ei.IsZero() {
			ei = past[i].Slot.StartTime
		}
		if ej.IsZero() {
			ej = past[j].Slot.StartTime
		}
		return ej.Before(ei)
	})
	return
}

// filterRecordsByStatus фильтрует записи по статусу
func (s *Service) filterRecordsByStatus(records []mymodels.Record, status string) []mymodels.Record {
	if status == "" {
		return records // Возвращаем все записи если статус не указан
	}

	var filtered []mymodels.Record
	for _, record := range records {
		if record.Status == status {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

func pagesCount(total, limit int) int {
	if limit <= 0 {
		return 1
	}
	pages := total / limit
	if total%limit != 0 {
		pages++
	}
	if pages == 0 {
		pages = 1
	}
	return pages
}
