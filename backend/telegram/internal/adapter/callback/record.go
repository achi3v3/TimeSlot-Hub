package callback

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"telegram-bot/internal/handlers/components"
	"telegram-bot/internal/handlers/message"
	record "telegram-bot/internal/handlers/record"
	"telegram-bot/internal/utils"
	mymodels "telegram-bot/pkg/models"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"
)

func (h *CallBackHandler) Records() {
	callbackQuery := h.update.CallbackQuery
	callbackData := callbackQuery.Data
	parts := strings.Split(callbackData, "/")
	if len(parts) >= 3 {
		status := parts[1]
		if status == "all" {
			status = ""
		}
		pageStr := parts[2]
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			page = 1
		}
		// При пагинации записей — редактируем конкретное сообщение
		svc := record.NewService(h.b, logrus.New())
		svc.EditUserRecordsPage(h.ctx, h.userID, h.userID, status, page, h.messageID)
		h.answerCallBackQuery(fmt.Sprintf("Стр. %d", page), false)
	}
}

func (h *CallBackHandler) RecordsTime() {
	callbackQuery := h.update.CallbackQuery
	callbackData := callbackQuery.Data
	parts := strings.Split(callbackData, "/")
	if len(parts) >= 3 {
		mode := parts[1]
		status := "all"
		page := 1

		// Определяем количество параметров и их значения
		if len(parts) >= 4 {
			// Формат: records_time/{mode}/{status}/{page}
			status = parts[2]
			if p, err := strconv.Atoi(parts[3]); err == nil {
				page = p
			}
		} else if len(parts) >= 3 {
			// Формат: records_time/{mode}/{page} (старый формат для обратной совместимости)
			if p, err := strconv.Atoi(parts[2]); err == nil {
				page = p
			}
		}

		svc := record.NewService(h.b, logrus.New())
		if mode == "chooser" {
			// Вернуться к выбору пула с сохранением статуса
			var text string
			var futureCallback, pastCallback string

			if status != "" && status != "all" {
				statusText := "все записи"
				switch status {
				case "confirm":
					statusText = "подтвержденные записи"
				case "reject":
					statusText = "отклоненные записи"
				case "pending":
					statusText = "записи в ожидании"
				}
				text = fmt.Sprintf("%s<b>Мои записи (%s)</b>\n<i>Выберите пул записей для просмотра</i>", components.Header(), statusText)
				futureCallback = fmt.Sprintf("records_time/future/%s/1", status)
				pastCallback = fmt.Sprintf("records_time/past/%s/1", status)
			} else {
				text = fmt.Sprintf("%s<b>Мои записи</b>\n<i>Выберите пул записей для просмотра</i>", components.Header())
				futureCallback = "records_time/future/all/1"
				pastCallback = "records_time/past/all/1"
			}

			kb := &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: "Будущие", CallbackData: futureCallback}},
				{{Text: "Прошедшие", CallbackData: pastCallback}},
			}}
			_ = messageEditor.EditUserRecords(h.ctx, h.b, h.userID, "chooser", 1, text, kb)
		} else {
			svc.EditUserRecordsTimePage(h.ctx, h.userID, h.userID, mode, status, page, h.messageID)
		}
		h.answerCallBackQuery("Обновлено", false)
	}
}
func (h *CallBackHandler) CheckUserAuth(userID int64) bool {
	_, exist := h.client.CheckAuth(h.ctx, userID)
	if !exist {
		msgText := fmt.Sprintf(
			"%s"+
				"<blockquote> ℹ️ Для начала использования бота, вам необходимо зарегистрироваться!\n\nДля регистрации /start</blockquote>",
			components.Header(),
		)
		h.b.SendMessage(h.ctx, &bot.SendMessageParams{
			ChatID:    userID,
			ParseMode: models.ParseModeHTML,
			Text:      msgText,
		})
		return false
	}
	return true
}

func (h *CallBackHandler) BookMove() {
	callbackQuery := h.update.CallbackQuery
	callbackData := callbackQuery.Data
	parts := strings.Split(callbackData, "/")
	if len(parts) >= 2 {
		slotIDStr := parts[1]
		slotID, err := strconv.Atoi(slotIDStr)
		if err != nil {
			log.Printf("Invalid slot ID in booking callback: %s", slotIDStr)
			return
		}
		if !h.CheckUserAuth(h.userID) {
			log.Printf("User not authorizated: %s", slotIDStr)
			return
		}
		// Получаем пользователя (для client_id)
		user, err := h.client.GetUserByTelegramID(h.ctx, h.userID)
		if err != nil || user == nil {
			log.Printf("GetUserByTelegramID failed: %v", err)
			// Редактируем сообщение, на котором была нажата кнопка
			if h.messageID != 0 {
				log.Printf("Editing message %d for user %d - user check error", h.messageID, h.userID)
				messageEditor.EditSpecificMessage(h.ctx, h.b, h.userID, h.messageID, msgErrorWithCheckUser, nil)
			} else {
				log.Printf("MessageID is 0, sending new message to user %d - user check error", h.userID)
				h.b.SendMessage(h.ctx, &bot.SendMessageParams{
					ChatID:    h.userID,
					Text:      msgErrorWithCheckUser,
					ParseMode: models.ParseModeHTML,
				})
			}
			h.answerCallBackQuery("Ошибка: пользователь не найден", false)
			return
		}
		// Получаем информацию о слоте для отображения деталей
		slot, ok := h.client.GetSlotByID(h.ctx, uint(slotID))
		if !ok {
			log.Printf("GetSlotByID failed for slot: %d", slotID)
			errorText := fmt.Sprintf("%s⚠️ Ошибка\n<i>Не удалось получить информацию о слоте</i>", components.Header())
			if h.messageID != 0 {
				log.Printf("Editing message %d for user %d - slot info error", h.messageID, h.userID)
				messageEditor.EditSpecificMessage(h.ctx, h.b, h.userID, h.messageID, errorText, nil)
			} else {
				log.Printf("MessageID is 0, sending new message to user %d - slot info error", h.userID)
				h.b.SendMessage(h.ctx, &bot.SendMessageParams{
					ChatID:    h.userID,
					Text:      errorText,
					ParseMode: models.ParseModeHTML,
				})
			}
			h.answerCallBackQuery("Ошибка получения информации о слоте", false)
			return
		}

		// Формируем запрос на создание записи
		req := mymodels.Record{
			SlotID:   uint(slotID),
			ClientID: user.ID,
			Status:   "pending",
		}
		msg, ok := h.client.CreateRecord(h.ctx, req)
		if !ok {
			log.Printf("CreateRecord failed: %s", msg)
			// Редактируем сообщение, на котором была нажата кнопка
			errorText := fmt.Sprintf("%s⚠️ Ошибка\n<i>Не удалось создать запись: %s</i>", components.Header(), msg)
			if h.messageID != 0 {
				log.Printf("Editing message %d for user %d - create record error", h.messageID, h.userID)
				messageEditor.EditSpecificMessage(h.ctx, h.b, h.userID, h.messageID, errorText, nil)
			} else {
				log.Printf("MessageID is 0, sending new message to user %d - create record error", h.userID)
				h.b.SendMessage(h.ctx, &bot.SendMessageParams{
					ChatID:    h.userID,
					Text:      errorText,
					ParseMode: models.ParseModeHTML,
				})
			}
			h.answerCallBackQuery("Ошибка создания записи", false)
			return
		}

		// Форматируем время с учетом таймзоны мастера
		date := utils.FormatDateInLocation(slot.MasterTimezone, slot.StartTime)
		startTime := utils.FormatTimeOnlyInLocation(slot.MasterTimezone, slot.StartTime)
		endTime := utils.FormatTimeOnlyInLocation(slot.MasterTimezone, slot.EndTime)

		// Получаем метку таймзоны для отображения (обязательно таймзона мастера)
		tzLabel := slot.MasterTimezone
		if tzLabel == "" {
			tzLabel = "Europe/Moscow" // на московскую таймзону, если таймзона не указана
		}

		// Получаем смещение таймзоны для отображения
		tzOffset := utils.GetTimezoneOffset(tzLabel)

		// Формируем время с таймзоной (таймзона только у времени, не у даты)
		timeWithTZ := fmt.Sprintf("%s - %s (TZ: %s %s)", startTime, endTime, tzLabel, tzOffset)

		// Формируем детальное сообщение о записи
		confirmText := fmt.Sprintf("%s✅ <b>Вы успешно записались!</b>\n\n"+
			"<b>Детали записи:</b>\n"+
			"<blockquote>"+
			"<b>Мастер:</b> <code>%s %s</code>\n"+
			"<b>Услуга:</b> <code>%s</code>\n"+
			"<b>Дата:</b> <code>%s</code>\n"+
			"<b>Время:</b> <code>%s</code>\n"+
			"<b>Длительность:</b> <code>%d мин.</code>\n"+
			"<b>Стоимость:</b> <code>%.0f руб.</code>\n"+
			"<b>Статус:</b> <code>⏳ Ожидает подтверждения</code>\n"+
			"</blockquote>\n\n"+
			"<i>Ожидайте подтверждения записи от мастера</i>\n\n"+
			"<i>Для просмотра всех ваших записей введите команду /allrecords</i>",
			components.Header(),
			slot.MasterName, slot.MasterSurname,
			slot.ServiceName,
			date,
			timeWithTZ,
			slot.ServiceDuration,
			slot.ServicePrice)

		// Удаляем исходное сообщение и отправляем новое (как в TryConfirmLogin)
		if h.messageID != 0 {
			log.Printf("Deleting message %d and sending new confirmation to user %d", h.messageID, h.userID)
			_, _ = h.b.DeleteMessage(h.ctx, &bot.DeleteMessageParams{ChatID: h.userID, MessageID: h.messageID})
		}
		h.b.SendMessage(h.ctx, &bot.SendMessageParams{
			ChatID:    h.userID,
			Text:      confirmText,
			ParseMode: models.ParseModeHTML,
		})
		// Удаляем состояние сообщения с деталями слота после успешного бронирования
		messageEditor.RemoveMessageState(h.userID, "slot_details")
		h.answerCallBackQuery("✅ Заявка отправлена", false)
	}
}

func (h *CallBackHandler) AllRecordsTime() {
	callbackQuery := h.update.CallbackQuery
	callbackData := callbackQuery.Data
	parts := strings.Split(callbackData, "/")
	if len(parts) >= 3 {
		mode := parts[1]
		status := "all"
		page := 1

		// Определяем количество параметров и их значения
		if len(parts) >= 4 {
			// Формат: all_records_time/{mode}/{status}/{page}
			status = parts[2]
			if p, err := strconv.Atoi(parts[3]); err == nil {
				page = p
			}
		}

		svc := record.NewService(h.b, logrus.New())
		if mode == "chooser" {
			// Вернуться к выбору времени для всех записей
			text := fmt.Sprintf("%s<b>Все мои записи</b>\n<i>Выберите пул записей для просмотра</i>", components.Header())
			kb := &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: "Будущие", CallbackData: "all_records_time/future/all/1"}},
				{{Text: "Прошедшие", CallbackData: "all_records_time/past/all/1"}},
			}}
			_ = messageEditor.EditUserRecords(h.ctx, h.b, h.userID, "chooser", 1, text, kb)
		} else {
			// Показать все записи по времени
			svc.EditUserRecordsTimePage(h.ctx, h.userID, h.userID, mode, status, page, h.messageID)
		}
		h.answerCallBackQuery("Обновлено", false)
	}
}

func (h *CallBackHandler) RecordAction() {
	callbackQuery := h.update.CallbackQuery
	callbackData := callbackQuery.Data
	parts := strings.Split(callbackData, "/")
	if len(parts) >= 3 {
		action := parts[1] // confirm или reject
		recordIDStr := parts[2]
		userID := callbackQuery.From.ID

		// Парсим recordID
		recordID, err := strconv.ParseUint(recordIDStr, 10, 0)
		fmt.Println(recordID)
		if err != nil {
			log.Printf("Invalid record ID in record_action callback: %s", recordIDStr)
			h.answerCallBackQuery("Ошибка: неверный ID записи", false)
			return
		}

		// Отправляем запрос на изменение статуса записи
		status := action
		msg, ok := h.client.UpdateRecordStatus(h.ctx, uint(recordID), status)
		if !ok {
			log.Printf("UpdateRecordStatus failed: %s", msg)
			h.answerCallBackQuery("Ошибка обновления статуса", false)
			return
		}

		var actionText string
		var emoji string
		if action == "confirm" {
			actionText = "подтверждена"
			emoji = "✅"
		} else {
			actionText = "отклонена"
			emoji = "❌"
		}

		// Обновляем сообщение с результатом действия
		newText := fmt.Sprintf("%s🆕 Новая запись\n<b>Запись %s</b>\n\n%s <b>Запись %s</b>",
			components.Header(),
			recordIDStr,
			emoji,
			actionText)

		// Убираем кнопки и показываем результат
		keyboard := &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{}}

		// Редактируем сообщение, на котором была нажата кнопка
		messageEditor.EditSpecificMessage(h.ctx, h.b, userID, h.messageID, newText, keyboard)

		// Используем AnswerCallbackQuery для уведомления пользователя
		h.answerCallBackQuery(fmt.Sprintf("Запись %s", actionText), true)
	}
}

// AccountDeletion обрабатывает callback'и для удаления аккаунта
func (h *CallBackHandler) AccountDeletion() {
	parts := strings.Split(h.query, "/")
	if len(parts) < 2 {
		h.answerCallBackQuery("Ошибка в данных", true)
		return
	}

	action := parts[1]
	userID := h.userID

	switch action {
	case "cancel":
		// Отменяем удаление аккаунта
		messageEditor := message.NewMessageEditor()
		newText := "❌ <b>Удаление аккаунта отменено</b>\n\nВаш аккаунт остается активным. Если у вас есть вопросы, обратитесь в поддержку."

		// Убираем кнопки
		keyboard := &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{}}

		// Редактируем сообщение
		messageEditor.EditSpecificMessage(h.ctx, h.b, userID, h.messageID, newText, keyboard)
		h.answerCallBackQuery("Удаление отменено", false)

	case "confirm":
		if len(parts) < 3 {
			h.answerCallBackQuery("Ошибка: не указан ID пользователя", true)
			return
		}

		userUUID := parts[2]

		// Отправляем запрос на подтверждение удаления в бэкенд
		err := h.client.ConfirmAccountDeletion(userUUID)
		if err != nil {
			log.Printf("AccountDeletion confirm error: %v", err)
			h.answerCallBackQuery("Ошибка при удалении аккаунта", true)
			return
		}

		// Показываем сообщение об успешном удалении
		messageEditor := message.NewMessageEditor()
		newText := "✅ <b>Аккаунт успешно удален</b>\n\nВсе ваши данные были безвозвратно удалены из системы. Спасибо за использование нашего сервиса!"

		// Убираем кнопки
		keyboard := &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{}}

		// Редактируем сообщение
		messageEditor.EditSpecificMessage(h.ctx, h.b, userID, h.messageID, newText, keyboard)
		h.answerCallBackQuery("Аккаунт удален", false)

	default:
		h.answerCallBackQuery("Неизвестное действие", true)
	}
}
