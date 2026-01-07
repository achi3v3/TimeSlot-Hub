package shared

import (
	"context"
	"fmt"
	"sort"
	"strings"
	adapter "telegram-bot/internal/adapter/backendapi"
	"telegram-bot/internal/app/formatter"
	"telegram-bot/internal/config"
	"telegram-bot/internal/handlers/components"
	"telegram-bot/internal/handlers/message"
	"telegram-bot/internal/logger"
	mymodels "telegram-bot/pkg/models"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var (
	cfg           = config.Load()
	log           = logger.New()
	client        = adapter.New(cfg.BackendBaseURL, log)
	messageEditor = message.NewMessageEditor()
)
var (
	msgErrorWithGetSlots   = fmt.Sprintf("%s⚠️ Ошибка\n<i>Произошла ошибка при получении слотов</i>", components.Header())
	msgErrorWithPagination = fmt.Sprintf("%s⚠️ Ошибка\n<i>Произошла ошибка при пагинации данных</i>", components.Header())
	msgNotSlots            = fmt.Sprintf("%sℹ️ Сообщение\n<i>Нет активных слотов</i>", components.Header())
	msgErrorWithKeyboard   = fmt.Sprintf("%s⚠️ Ошибка\n<i>Произошла ошибка при обработки кнопок</i>", components.Header())
)

func IsValidPrivateMessage(update *models.Update) bool {
	return formatter.IsValidPrivateMessage(update)
}

func ExtractUserID(update *models.Update) int64 {
	return formatter.ExtractUserID(update)
}

func SendGetUserLink(ctx context.Context, b *bot.Bot, userID int64) {
	user, err := client.GetUserByTelegramID(ctx, userID)
	if err != nil {
		log.Errorf("Failed to get slots for masterID: %d", userID)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    userID,
			ParseMode: models.ParseModeHTML,
			Text:      msgErrorWithGetSlots,
		})
		return
	}
	msg := fmt.Sprintf(
		"%s"+
			"Пользователь: <code>%s %s</code>\n\n"+
			"Мои услуги:\n"+
			"%s"+
			"Вы можете просмотреть:\n"+
			`<a href="%s">Моё расписание в Телеграме</a>`+
			"\n"+
			`<a href="%s/master/%s">Моё расписание на сайте</a>`+
			"\n\n",
		components.Header(),
		user.FirstName,
		user.Surname,
		PaginationUserServices(user.Services),
		buildBotStartLink(cfg.BotLink, userID),
		strings.TrimSuffix(cfg.PublicSiteURL, "/"),
		user.ID)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ParseMode: models.ParseModeHTML,
		ChatID:    userID,
		Text:      msg,
	})
}
func PaginationUserServices(services []mymodels.Service) string {
	if len(services) == 0 {
		return "<i>Услуги отсутствуют</i>"
	}
	var s strings.Builder
	for _, service := range services {
		s.WriteString("<blockquote><code>" + service.Name + "</code></blockquote>" + "\n")
	}
	return s.String()
}
func SendGetUserSlots(ctx context.Context, b *bot.Bot, userID int64, masterID int64) {
	slots, ok := client.GetSlotsByTelegramID(ctx, masterID)
	if !ok {
		log.Errorf("Failed to get slots for masterID: %d", masterID)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    userID,
			ParseMode: models.ParseModeHTML,
			Text:      msgErrorWithGetSlots,
		})
		return
	}
	if len(slots) == 0 {
		log.Infof("No slots found for masterID: %d", masterID)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    userID,
			ParseMode: models.ParseModeHTML,
			Text:      msgNotSlots,
		})
		return
	}

	log.Infof("Found %d slots for masterID: %d", len(slots), masterID)

	// Показываем меню выбора времени (Будущие/Прошедшие)
	text := fmt.Sprintf("%s<b>Мои слоты</b>\n<i>Выберите пул слотов для просмотра</i>", components.Header())
	keyboard := &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
		{{Text: "Будущие", CallbackData: fmt.Sprintf("slots_time/future/%d/1", masterID)}},
		{{Text: "Прошедшие", CallbackData: fmt.Sprintf("slots_time/past/%d/1", masterID)}},
	}}

	// Отправляем новое сообщение и устанавливаем состояние
	sent, err := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: userID, ParseMode: models.ParseModeHTML, Text: text, ReplyMarkup: keyboard})
	if err == nil {
		sm := message.GetStateManager()
		sm.RemoveMessageState(userID, userID, "slots_time_chooser")
		sm.SetMessageState(userID, userID, sent.ID, "slots_time_chooser", map[string]interface{}{
			"masterID": masterID,
			"timeType": "chooser",
			"page":     1,
		})
	}
}

func buildBotStartLink(botLink string, userID int64) string {
	base := strings.TrimSpace(botLink)
	if base == "" {
		return ""
	}
	separator := "?"
	if strings.Contains(base, "?") {
		separator = "&"
	}
	return fmt.Sprintf("%s%vstart=%d", base, separator, userID)
}

// SendFutureSlotsForClient отправляет только будущие слоты для клиента
func SendFutureSlotsForClient(ctx context.Context, b *bot.Bot, userID int64, masterID int64, page int, messageID int) {
	if page <= 0 {
		page = 1
	}
	slots, ok := client.GetSlotsByTelegramID(ctx, masterID)
	if !ok {
		log.Errorf("Failed to get slots for masterID: %d", masterID)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    userID,
			ParseMode: models.ParseModeHTML,
			Text:      msgErrorWithGetSlots,
		})
		return
	}
	if len(slots) == 0 {
		log.Infof("No slots found for masterID: %d", masterID)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    userID,
			ParseMode: models.ParseModeHTML,
			Text:      msgNotSlots,
		})
		return
	}

	log.Infof("Found %d slots for masterID: %d", len(slots), masterID)

	// Фильтруем только будущие слоты
	filteredSlots := filterSlotsByTime(slots, "future")

	if len(filteredSlots) == 0 {
		text := fmt.Sprintf("%s<b>Слоты мастера</b>\n<i>У мастера пока нет доступных слотов</i>", components.Header())
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    userID,
			ParseMode: models.ParseModeHTML,
			Text:      text,
		})
		return
	}

	// Создаем пагинированные слоты для будущих слотов
	paginationData := formatter.CreatePaginatedSlots(filteredSlots, "", page, 10, true)

	if paginationData == nil {
		log.Errorf("Failed to create pagination data for future slots")
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    userID,
			Text:      msgErrorWithPagination,
			ParseMode: models.ParseModeHTML,
		})
		return
	}
	if paginationData.CurrentPage < 1 {
		paginationData.CurrentPage = 1
	}
	// Проставим masterID в пагинационные данные
	paginationData.MasterInfo.TelegramID = masterID
	paginationData.IsClientView = true
	inlineKeyboard, text := formatter.CreatePaginatedInlineKeyboard(paginationData)

	if inlineKeyboard == nil {
		log.Errorf("Failed to create inline keyboard for pagination data")
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    userID,
			Text:      msgErrorWithKeyboard,
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	text = fmt.Sprintf("%s<b>Доступные слоты мастера</b>\n\n%s", components.Header(), text)

	// Редактируем конкретное сообщение, если messageID > 0, иначе отправляем новое
	if messageID > 0 {
		err := messageEditor.EditSpecificMessage(ctx, b, userID, messageID, text, inlineKeyboard)
		if err != nil {
			log.Errorf("Failed to edit future slots message: %v", err)
		}
	} else {
		// Отправляем новое сообщение
		sent, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      userID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: inlineKeyboard,
		})
		if err != nil {
			log.Errorf("Failed to send future slots message: %v", err)
			return
		}

		// Устанавливаем состояние для нового сообщения
		sm := message.GetStateManager()
		sm.RemoveMessageState(userID, userID, "client_slots")
		sm.SetMessageState(userID, userID, sent.ID, "client_slots", map[string]interface{}{
			"masterID": masterID,
			"timeType": "future",
			"page":     page,
		})
	}
}

// SendPaginatedFutureSlotsForClient paginates future-only slots by target date for client view
func SendPaginatedFutureSlotsForClient(ctx context.Context, b *bot.Bot, userID, masterID int64, targetDate string, page int, messageID int) {
	slots, ok := client.GetSlotsByTelegramID(ctx, masterID)
	if !ok {
		log.Errorf("Failed to get slots for masterID: %d", masterID)
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: userID, ParseMode: models.ParseModeHTML, Text: msgErrorWithGetSlots})
		return
	}
	if len(slots) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: userID, ParseMode: models.ParseModeHTML, Text: msgNotSlots})
		return
	}
	filteredSlots := filterSlotsByTime(slots, "future")
	if len(filteredSlots) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: userID, ParseMode: models.ParseModeHTML, Text: msgNotSlots})
		return
	}
	paginationData := formatter.CreatePaginatedSlots(filteredSlots, targetDate, page, 10, true)
	if paginationData == nil {
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: userID, Text: msgErrorWithPagination})
		return
	}
	paginationData.MasterInfo.TelegramID = masterID
	paginationData.IsClientView = true
	inlineKeyboard, text := formatter.CreatePaginatedInlineKeyboard(paginationData)
	if inlineKeyboard == nil {
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: userID, Text: msgErrorWithKeyboard})
		return
	}
	text = fmt.Sprintf("%s<b>Доступные слоты мастера</b>\n\n%s", components.Header(), text)
	if messageID > 0 {
		if err := messageEditor.EditSpecificMessage(ctx, b, userID, messageID, text, inlineKeyboard); err != nil {
			log.Errorf("Failed to edit future slots message: %v", err)
		}
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: userID, Text: text, ParseMode: models.ParseModeHTML, ReplyMarkup: inlineKeyboard})
	}
}

// SendPaginatedSlots отправляет пагинированные слоты для конкретной даты и страницы
func SendPaginatedSlots(ctx context.Context, b *bot.Bot, userID, masterID int64, targetDate string, page int, messageID int) {
	slots, ok := client.GetSlotsByTelegramID(ctx, masterID)
	if !ok {
		log.Errorf("Failed to get slots for masterID: %d", masterID)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    userID,
			ParseMode: models.ParseModeHTML,
			Text:      msgErrorWithGetSlots,
		})
		return
	}
	if len(slots) == 0 {
		log.Infof("No slots found for masterID: %d", masterID)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    userID,
			Text:      msgNotSlots,
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	log.Infof("Found %d slots for masterID: %d, requesting date: %s, page: %d", len(slots), masterID, targetDate, page)

	// Создаем пагинированные слоты для конкретной даты и страницы
	paginationData := formatter.CreatePaginatedSlots(slots, targetDate, page, 10, false)

	if paginationData == nil {
		log.Errorf("Failed to create pagination data for date %s, page %d", targetDate, page)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    userID,
			Text:      msgErrorWithPagination,
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	// Проставим masterID в пагинационные данные для корректного формирования callback
	paginationData.MasterInfo.TelegramID = masterID
	inlineKeyboard, text := formatter.CreatePaginatedInlineKeyboard(paginationData)

	if inlineKeyboard == nil {
		log.Errorf("Failed to create inline keyboard for pagination data")
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    userID,
			Text:      msgErrorWithKeyboard,
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	log.Infof("Created paginated keyboard for date %s, page %d/%d with %d buttons",
		paginationData.CurrentDate, paginationData.CurrentPage, paginationData.TotalPages, len(inlineKeyboard.InlineKeyboard))

	// Редактируем конкретное сообщение, если messageID > 0, иначе отправляем новое
	if messageID > 0 {
		err := messageEditor.EditSpecificMessage(ctx, b, userID, messageID, text, inlineKeyboard)
		if err != nil {
			log.Errorf("Failed to edit paginated slots message: %v", err)
		}
	} else {
		// Отправляем новое сообщение
		sent, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      userID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: inlineKeyboard,
		})
		if err != nil {
			log.Errorf("Failed to send paginated slots message: %v", err)
		} else {
			// Устанавливаем состояние для нового сообщения
			sm := message.GetStateManager()
			sm.SetMessageState(userID, userID, sent.ID, "slots_pagination", map[string]interface{}{
				"masterID":   masterID,
				"targetDate": targetDate,
				"page":       page,
			})
		}
	}
}

// SendSlotsByTime отправляет слоты, отфильтрованные по времени (будущие/прошедшие)
func SendSlotsByTime(ctx context.Context, b *bot.Bot, userID, masterID int64, timeType string, page int, messageID int) {
	slots, ok := client.GetSlotsByTelegramID(ctx, masterID)
	if !ok {
		log.Errorf("Failed to get slots for masterID: %d", masterID)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    userID,
			ParseMode: models.ParseModeHTML,
			Text:      msgErrorWithGetSlots,
		})
		return
	}
	if len(slots) == 0 {
		log.Infof("No slots found for masterID: %d", masterID)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    userID,
			Text:      msgNotSlots,
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	log.Infof("Found %d slots for masterID: %d, filtering by time: %s, page: %d", len(slots), masterID, timeType, page)

	// Фильтруем слоты по времени
	filteredSlots := filterSlotsByTime(slots, timeType)

	if len(filteredSlots) == 0 {
		timeText := "будущих"
		if timeType == "past" {
			timeText = "прошедших"
		}
		text := fmt.Sprintf("%s<b>Мои слоты (%s)</b>\n<i>У вас пока нет %s слотов</i>", components.Header(), timeText, timeText)
		keyboard := &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "◀️ К выбору", CallbackData: fmt.Sprintf("slots_time/chooser/%d", masterID)}},
		}}

		// Отправляем новое сообщение вместо редактирования
		sent, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      userID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: keyboard,
		})
		if err != nil {
			log.Errorf("Failed to send empty slots message: %v", err)
			return
		}

		// Устанавливаем состояние для нового сообщения
		sm := message.GetStateManager()
		sm.RemoveMessageState(userID, userID, "slots_time_chooser")
		sm.SetMessageState(userID, userID, sent.ID, "slots_time_chooser", map[string]interface{}{
			"masterID": masterID,
			"timeType": timeType,
			"page":     page,
		})
		return
	}

	// Создаем пагинированные слоты для отфильтрованного списка
	paginationData := formatter.CreatePaginatedSlots(filteredSlots, "", page, 10, false)

	if paginationData == nil {
		log.Errorf("Failed to create pagination data for filtered slots")
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    userID,
			Text:      msgErrorWithPagination,
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	// Проставим masterID в пагинационные данные
	paginationData.MasterInfo.TelegramID = masterID
	inlineKeyboard, text := formatter.CreatePaginatedInlineKeyboard(paginationData)

	if inlineKeyboard == nil {
		log.Errorf("Failed to create inline keyboard for pagination data")
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    userID,
			Text:      msgErrorWithKeyboard,
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	// Добавляем кнопку возврата к выбору времени
	timeText := "будущих"
	if timeType == "past" {
		timeText = "прошедших"
	}
	text = fmt.Sprintf("%s<b>Мои слоты (%s)</b>\n\n%s", components.Header(), timeText, text)

	// Добавляем кнопку "К выбору" в клавиатуру
	if inlineKeyboard.InlineKeyboard != nil {
		inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, []models.InlineKeyboardButton{
			{Text: "◀️ К выбору", CallbackData: fmt.Sprintf("slots_time/chooser/%d", masterID)},
		})
	}

	log.Infof("Created paginated keyboard for %s slots, page %d/%d", timeType, paginationData.CurrentPage, paginationData.TotalPages)

	// Редактируем конкретное сообщение, если messageID > 0, иначе отправляем новое
	if messageID > 0 {
		err := messageEditor.EditSpecificMessage(ctx, b, userID, messageID, text, inlineKeyboard)
		if err != nil {
			log.Errorf("Failed to edit slots by time message: %v", err)
		}
	} else {
		// Отправляем новое сообщение
		sent, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      userID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: inlineKeyboard,
		})
		if err != nil {
			log.Errorf("Failed to send slots by time message: %v", err)
			return
		}

		// Устанавливаем состояние для нового сообщения
		sm := message.GetStateManager()
		sm.RemoveMessageState(userID, userID, "slots_time_chooser")
		sm.SetMessageState(userID, userID, sent.ID, "slots_time_chooser", map[string]interface{}{
			"masterID": masterID,
			"timeType": timeType,
			"page":     page,
		})
	}
}

// filterSlotsByTime фильтрует слоты по времени (будущие/прошедшие)
func filterSlotsByTime(slots []mymodels.SlotResponse, timeType string) []mymodels.SlotResponse {
	now := time.Now()
	threshold := time.Hour // 1 час просрочки

	var filtered []mymodels.SlotResponse
	for _, slot := range slots {
		start := slot.StartTime
		end := slot.EndTime
		if end.IsZero() && !start.IsZero() && slot.ServiceDuration > 0 {
			end = start.Add(time.Duration(slot.ServiceDuration) * time.Minute)
		}

		isPast := false
		if !end.IsZero() {
			isPast = end.Add(threshold).Before(now)
		} else if !start.IsZero() {
			isPast = start.Add(threshold).Before(now)
		}

		if timeType == "past" && isPast {
			filtered = append(filtered, slot)
		} else if timeType == "future" && !isPast {
			filtered = append(filtered, slot)
		}
	}

	// Сортировка: будущие по возрастанию start, прошедшие по убыванию end
	if timeType == "future" {
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].StartTime.Before(filtered[j].StartTime)
		})
	} else if timeType == "past" {
		sort.Slice(filtered, func(i, j int) bool {
			ei := filtered[i].EndTime
			ej := filtered[j].EndTime
			if ei.IsZero() {
				ei = filtered[i].StartTime
			}
			if ej.IsZero() {
				ej = filtered[j].StartTime
			}
			return ej.Before(ei)
		})
	}

	return filtered
}
