package callback

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"telegram-bot/internal/handlers/components"
	"telegram-bot/internal/handlers/message"
	"telegram-bot/internal/handlers/shared"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var (
	msgNotRegister           = fmt.Sprintf("%s⚠️ Ошибка\n<i>🔒 Пожалуйста, сначала зарегистрируйтесь через /start</i>", components.Header())
	msgIdDoesntMatch         = fmt.Sprintf("%s⚠️ Ошибка\n<blockquote><i>TelegramID контакта не совпадает с вашим ID, проверьте, что вы передаёте свой контакт!</i></blockquote>", components.Header())
	empty                    = fmt.Sprintf("%s", components.Header())
	msgErrorWithConfirmLogin = fmt.Sprintf("%s⚠️ Ошибка\n<i>🔒 Не удалось подтвердить вход</i>", components.Header())
	SuccessConfirmLogin      = fmt.Sprintf("%s✔️ Успешно\n<i>Вход подтверждён!</i>", components.Header())
	msgErrorWithCheckUser    = fmt.Sprintf("%s⚠️ Ошибка\n<i>🔒Не удалось определить пользователя. Пройдите регистрацию /start если вы ещё не зарегистрировались</i>", components.Header())
	SuccessCreateRecord      = fmt.Sprintf("%s✔️ Успешно\n<i>Заявка на запись отправлена. Ожидайте ответа на заявку от мастера.", components.Header())
	messageEditor            = message.NewMessageEditor()
)

func (h *CallBackHandler) answerCallBackQuery(text string, showAlert bool) {
	callbackQuery := h.update.CallbackQuery
	h.b.AnswerCallbackQuery(h.ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callbackQuery.ID,
		Text:            text,
		ShowAlert:       showAlert,
	})
}
func answerCallBackQuery(ctx context.Context, b *bot.Bot, update *models.Update, text string, showAlert bool) {
	callbackQuery := update.CallbackQuery
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callbackQuery.ID,
		Text:            text,
		ShowAlert:       showAlert,
	})
}
func (h *CallBackHandler) DateMove() {
	callbackQuery := h.update.CallbackQuery
	callbackData := callbackQuery.Data
	parts := strings.Split(callbackData, "/")
	// date/{masterTelegramID}/{date}/{page}
	if len(parts) >= 4 {
		masterTelegramStr := parts[1]
		targetDate := parts[2]
		pageStr := parts[3]
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			log.Printf("Invalid page number in callback: %s", pageStr)
			return
		}

		// masterID берём из callback data
		masterTelegramID, err := strconv.ParseInt(masterTelegramStr, 10, 64)
		if err != nil {
			log.Printf("Invalid master telegram id in callback: %s", masterTelegramStr)
			return
		}

		log.Printf("Navigating to date: %s, page: %d for user: %d, master: %d", targetDate, page, h.userID, masterTelegramID)

		// Навигация по датам будет редактировать сообщение пагинации
		shared.SendPaginatedSlots(h.ctx, h.b, h.userID, masterTelegramID, targetDate, page, h.messageID)

		// Отвечаем на callback query
		h.answerCallBackQuery(fmt.Sprintf("Переход к %s, страница %d", targetDate, page), false)
	}
}

func (h *CallBackHandler) DateMoveClient() {
	callbackQuery := h.update.CallbackQuery
	callbackData := callbackQuery.Data
	parts := strings.Split(callbackData, "/")
	// client_date/{masterTelegramID}/{date}/{page}
	if len(parts) >= 4 {
		masterTelegramStr := parts[1]
		targetDate := parts[2]
		pageStr := parts[3]
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			log.Printf("Invalid page number in callback: %s", pageStr)
			return
		}

		masterTelegramID, err := strconv.ParseInt(masterTelegramStr, 10, 64)
		if err != nil {
			log.Printf("Invalid master telegram id in callback: %s", masterTelegramStr)
			return
		}

		shared.SendPaginatedFutureSlotsForClient(h.ctx, h.b, h.userID, masterTelegramID, targetDate, page, h.messageID)
		h.answerCallBackQuery(fmt.Sprintf("Переход к %s, страница %d", targetDate, page), false)
	}
}

func formatFloat(num float64) string {
	// Форматируем с двумя знаками после запятой
	str := fmt.Sprintf("%.2f", num)
	// Заменяем точку на запятую
	return strings.Replace(str, ".", ",", -1)
}
