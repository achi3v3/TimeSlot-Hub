package callback

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"telegram-bot/internal/handlers/components"
	"telegram-bot/internal/handlers/shared"
	"telegram-bot/internal/utils"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *CallBackHandler) SlotMove() {
	callbackQuery := h.update.CallbackQuery
	callbackData := callbackQuery.Data
	parts := strings.Split(callbackData, "/")
	if len(parts) >= 2 {
		slotIDStr := parts[1]
		slotID, err := strconv.Atoi(slotIDStr)
		if err != nil {
			log.Printf("Invalid slot ID in callback: %s", slotIDStr)
			return
		}
		slot, ok := h.client.GetSlotByID(h.ctx, uint(slotID))
		fmt.Printf("%+v", slot)
		if !ok {
			log.Printf("Invalid get slot: %s", slotIDStr)
			return
		}
		log.Printf("User %d selected slot: %d", h.userID, slotID)

		var statusSlot string
		if !slot.IsBooked {
			statusSlot = "Свободен"
		} else {
			statusSlot = "Забронирован"
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

		// Формируем inline-кнопки действий
		buttons := [][]models.InlineKeyboardButton{}
		if !slot.IsBooked {
			buttons = append(buttons, []models.InlineKeyboardButton{{
				Text:         "📝 Записаться",
				CallbackData: fmt.Sprintf("book/%d", slotID),
			}})
		}

		// Добавляем кнопку "Назад к слотам"
		buttons = append(buttons, []models.InlineKeyboardButton{{
			Text:         "⬅️ Назад к слотам",
			CallbackData: fmt.Sprintf("back_to_slots/%d/%s/1", slot.MasterTelegramID, date),
		}})
		keyboard := &models.InlineKeyboardMarkup{InlineKeyboard: buttons}

		// Формируем время с таймзоной (таймзона только у времени, не у даты)
		timeWithTZ := fmt.Sprintf("%s — %s (TZ: %s %s)", startTime, endTime, tzLabel, tzOffset)
		slotDetailsText := fmt.Sprintf("%sПользователь:\n<code>%s %s</code>\n\nСлот:\n<blockquote><code>%s</code> [ %s ]\nДата: <code>%s</code>\nУслуга: <code>%s</code>\nДополнительно:\n %d мин. / %s руб.\n</blockquote>\n\n", components.Header(),
			slot.MasterName, slot.MasterSurname, timeWithTZ, statusSlot, date, slot.ServiceName, slot.ServiceDuration, formatFloat(slot.ServicePrice))

		// Редактируем сообщение, на котором была нажата кнопка
		if h.messageID != 0 {
			log.Printf("Editing message %d for user %d - slot details", h.messageID, h.userID)
			err = messageEditor.EditSpecificMessage(h.ctx, h.b, h.userID, h.messageID, slotDetailsText, keyboard)
			if err != nil {
				log.Printf("Failed to edit slot details message: %v", err)
			}
		} else {
			log.Printf("MessageID is 0, sending new message to user %d - slot details", h.userID)
			h.b.SendMessage(h.ctx, &bot.SendMessageParams{
				ChatID:      h.userID,
				Text:        slotDetailsText,
				ParseMode:   models.ParseModeHTML,
				ReplyMarkup: keyboard,
			})
		}

		// Отвечаем на callback query
		h.answerCallBackQuery("Выбран слот", false)
	}
}

func (h *CallBackHandler) BackToSlots() {
	callbackQuery := h.update.CallbackQuery
	callbackData := callbackQuery.Data
	parts := strings.Split(callbackData, "/")
	if len(parts) >= 4 {
		masterTelegramStr := parts[1]
		targetDate := parts[2]
		pageStr := parts[3]
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			log.Printf("Invalid page number in back_to_slots callback: %s", pageStr)
			return
		}

		masterTelegramID, err := strconv.ParseInt(masterTelegramStr, 10, 64)
		if err != nil {
			log.Printf("Invalid master telegram id in back_to_slots callback: %s", masterTelegramStr)
			return
		}

		log.Printf("Going back to slots for date: %s, page: %d for user: %d, master: %d", targetDate, page, h.userID, masterTelegramID)

		// Удаляем состояние деталей слота
		messageEditor.RemoveMessageState(h.userID, "slot_details")

		// Показываем слоты (это будет редактировать сообщение пагинации)
		shared.SendPaginatedSlots(h.ctx, h.b, h.userID, masterTelegramID, targetDate, page, h.messageID)
		h.answerCallBackQuery("Возврат к слотам", false)
	}
}

func (h *CallBackHandler) SlotsTime() {
	callbackQuery := h.update.CallbackQuery
	callbackData := callbackQuery.Data
	parts := strings.Split(callbackData, "/")
	if len(parts) >= 3 {
		mode := parts[1]
		masterIDStr := parts[2]
		page := 1

		// Определяем количество параметров и их значения
		if len(parts) >= 4 {
			// Формат: slots_time/{mode}/{masterID}/{page}
			if p, err := strconv.Atoi(parts[3]); err == nil {
				page = p
			}
		}

		masterID, err := strconv.ParseInt(masterIDStr, 10, 64)
		if err != nil {
			log.Printf("Invalid master ID in slots_time callback: %s", masterIDStr)
			return
		}

		if mode == "chooser" {
			// Вернуться к выбору времени
			text := fmt.Sprintf("%s<b>Мои слоты</b>\n<i>Выберите пул слотов для просмотра</i>", components.Header())
			kb := &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: "Будущие", CallbackData: fmt.Sprintf("slots_time/future/%d/1", masterID)}},
				{{Text: "Прошедшие", CallbackData: fmt.Sprintf("slots_time/past/%d/1", masterID)}},
			}}
			_ = messageEditor.EditSlotsPagination(h.ctx, h.b, h.userID, masterID, "chooser", 1, text, kb)
		} else {
			// Показать слоты по времени
			shared.SendSlotsByTime(h.ctx, h.b, h.userID, masterID, mode, page, h.messageID)
		}
		h.answerCallBackQuery("Обновлено", false)
	}
}
