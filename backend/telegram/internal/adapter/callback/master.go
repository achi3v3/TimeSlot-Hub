package callback

import (
	"log"
	"strconv"
	"strings"
	"telegram-bot/internal/handlers/master"

	"github.com/go-telegram/bot"
	"github.com/sirupsen/logrus"
)

// UpcomingPage обрабатывает пагинацию предстоящих записей мастера
func (h *CallBackHandler) UpcomingPage() {
	callbackQuery := h.update.CallbackQuery
	callbackData := callbackQuery.Data
	parts := strings.Split(callbackData, "/")

	if len(parts) < 3 {
		log.Printf("Invalid upcoming_page callback data: %s", callbackData)
		return
	}

	page, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Printf("Invalid page number in upcoming_page callback: %s", parts[1])
		return
	}

	telegramID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		log.Printf("Invalid telegram_id in upcoming_page callback: %s", parts[2])
		return
	}

	// Получаем записи с backend
	records, err := h.client.GetUpcomingRecordsByMasterTelegramID(h.ctx, telegramID)
	if err != nil {
		log.Printf("Failed to get upcoming records: %v", err)
		h.answerCallBackQuery("Ошибка при получении записей", true)
		return
	}

	// Создаем handler для форматирования
	masterHandler := master.NewHandler(logrus.New())

	// Форматируем и отправляем страницу
	const limit = 5
	text, totalPages := masterHandler.FormatUpcomingRecordsPage(records, page, limit, telegramID)
	keyboard := masterHandler.BuildUpcomingRecordsPagination(page, totalPages, telegramID)

	// Редактируем сообщение
	if h.messageID != 0 {
		_, err := h.b.EditMessageText(h.ctx, &bot.EditMessageTextParams{
			ChatID:      h.userID,
			MessageID:   h.messageID,
			Text:        text,
			ParseMode:   "HTML",
			ReplyMarkup: keyboard,
		})
		if err != nil {
			log.Printf("Failed to edit upcoming records message: %v", err)
		}
	}

	h.answerCallBackQuery("", false)
}
