package callback

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"telegram-bot/internal/config"
	"telegram-bot/internal/handlers/components"
	"telegram-bot/internal/handlers/shared"
	"telegram-bot/pkg/encrypt"
	mymodels "telegram-bot/pkg/models"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handler) ContactHandler() {
	contact := h.update.Message.Contact
	userID := h.update.Message.From.ID
	if contact.UserID != userID {
		h.b.SendMessage(h.ctx, &bot.SendMessageParams{
			ChatID:    userID,
			ParseMode: models.ParseModeHTML,
			Text:      msgIdDoesntMatch,
		})
		return
	}

	token, err := encrypt.GenerateToken(userID, contact.PhoneNumber)
	if err != nil {
		log.Printf("UniversalHandler: " + "error with generate token")
		return
	}

	userRequest := mymodels.UserRegister{
		Phone:      contact.PhoneNumber,
		TelegramID: userID,
		FirstName:  contact.FirstName,
		Surname:    contact.LastName,
		Token:      token,
		Active:     true,
	}
	var msgText string
	message, success := h.client.RegisterUser(h.ctx, userRequest)
	if !success {
		msgText = fmt.Sprintf(
			"%s"+"⚠️ Ошибка\n<i>Не удалось зарегистрироваться!</i>\n"+
				"Номер: <code>%s</code>\n"+
				"Имя: <code>%s %s</code>\n"+
				"<blockquote><i>Сообщение: %s\n</i></blockquote>", components.Header(), contact.PhoneNumber, contact.FirstName, contact.LastName, message)
	} else {
		msgText = fmt.Sprintf(
			"%s \n✔️ Успешно <i>Регистрация прошла!</i>\n"+
				"Номер: <code>%s</code>\n"+
				"Имя: <code>%s %s</code>\n\n"+
				`<a href='%s'>Перейти на сайт</a>`, components.Header(),
			contact.PhoneNumber, contact.FirstName, contact.LastName, strings.TrimSuffix(config.Load().PublicSiteURL, "/"))
	}

	h.b.SendMessage(h.ctx, &bot.SendMessageParams{
		ChatID:    userID,
		ParseMode: models.ParseModeHTML,
		Text:      msgText,
	})

	// После регистрации предлагаем выбрать таймзону
	timezoneKeyboard := [][]models.InlineKeyboardButton{
		{{Text: "Москва (Europe/Moscow, +03:00)", CallbackData: "tz/Europe/Moscow"}},
		{{Text: "Екатеринбург (Asia/Yekaterinburg, +05:00)", CallbackData: "tz/Asia/Yekaterinburg"}},
		{{Text: "Новосибирск (Asia/Novosibirsk, +07:00)", CallbackData: "tz/Asia/Novosibirsk"}},
		{{Text: "Омск (Asia/Omsk, +06:00)", CallbackData: "tz/Asia/Omsk"}},
		{{Text: "Самара (Europe/Samara, +04:00)", CallbackData: "tz/Europe/Samara"}},
	}
	h.b.SendMessage(h.ctx, &bot.SendMessageParams{
		ChatID:    userID,
		ParseMode: models.ParseModeHTML,
		Text:      "Выберите вашу таймзону:",
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: timezoneKeyboard,
		},
	})
}

func (h *CallBackHandler) TryConfirmLogin() {
	// Всегда отвечаем на callback, чтобы кнопка "крутилка" исчезала у пользователя
	h.answerCallBackQuery("Обрабатываю…", false)

	msg, ok := h.client.CheckAuth(h.ctx, h.userID)
	if !ok {
		log.Printf("!ok with CheckAuth: %v", msg)
		// Редактируем сообщение, на котором была нажата кнопка
		if h.messageID != 0 {
			log.Printf("Editing message %d for user %d - user not registered", h.messageID, h.userID)
			messageEditor.EditSpecificMessage(h.ctx, h.b, h.userID, h.messageID, msgNotRegister, nil)
		} else {
			log.Printf("MessageID is 0, sending new message to user %d - user not registered", h.userID)
			h.b.SendMessage(h.ctx, &bot.SendMessageParams{
				ChatID:    h.userID,
				Text:      msgNotRegister,
				ParseMode: models.ParseModeHTML,
			})
		}
		h.answerCallBackQuery("Ошибка авторизации", false)
		return
	}
	// fmt.Printf("%+v", h)
	msg, ok = h.client.ConfirmLogin(h.ctx, h.userID)
	if !ok {
		log.Printf("!ok with ConfirmLogin: %v", msg)
		// Редактируем сообщение, на котором была нажата кнопка
		if h.messageID != 0 {
			log.Printf("Editing message %d for user %d - confirm login error", h.messageID, h.userID)
			messageEditor.EditSpecificMessage(h.ctx, h.b, h.userID, h.messageID, msgErrorWithConfirmLogin, nil)
		} else {
			log.Printf("MessageID is 0, sending new message to user %d - confirm login error", h.userID)
			h.b.SendMessage(h.ctx, &bot.SendMessageParams{
				ChatID:    h.userID,
				Text:      msgErrorWithConfirmLogin,
				ParseMode: models.ParseModeHTML,
			})
		}
		h.answerCallBackQuery("Ошибка подтверждения входа", false)
		return
	}
	// Новое поведение: удаляем исходное сообщение и отправляем новое сообщение об успехе
	if h.messageID != 0 {
		_, _ = h.b.DeleteMessage(h.ctx, &bot.DeleteMessageParams{ChatID: h.userID, MessageID: h.messageID})
	}
	h.b.SendMessage(h.ctx, &bot.SendMessageParams{
		ChatID:    h.userID,
		Text:      SuccessConfirmLogin,
		ParseMode: models.ParseModeHTML,
	})
	h.answerCallBackQuery("Вход подтвержден", false)
}

func (h *CallBackHandler) SelectTimezone() {
	callbackQuery := h.update.CallbackQuery
	data := callbackQuery.Data // tz/<IANA>
	if !strings.HasPrefix(data, "tz/") {
		h.answerCallBackQuery("Неверный формат таймзоны", true)
		return
	}
	tz := strings.TrimPrefix(data, "tz/")
	if err := h.client.UpdateTimezoneInternal(h.ctx, h.userID, tz); err != nil {
		h.answerCallBackQuery("Не удалось сохранить таймзону", true)
		return
	}
	// Уведомляем пользователя и убираем инлайн-клавиатуру
	h.b.EditMessageText(h.ctx, &bot.EditMessageTextParams{
		ChatID:    h.userID,
		MessageID: callbackQuery.Message.Message.ID,
		Text:      fmt.Sprintf("Таймзона установлена: %s", tz),
		ParseMode: models.ParseModeHTML,
	})
	h.answerCallBackQuery("Сохранено", false)
}

func (h *CallBackHandler) ClientSlots() {
	callbackQuery := h.update.CallbackQuery
	callbackData := callbackQuery.Data
	parts := strings.Split(callbackData, "/")
	if len(parts) >= 3 {
		masterIDStr := parts[1]
		pageStr := parts[2]

		masterID, err := strconv.ParseInt(masterIDStr, 10, 64)
		if err != nil {
			log.Printf("Invalid master ID in client_slots callback: %s", masterIDStr)
			return
		}

		page, err := strconv.Atoi(pageStr)
		if err != nil {
			log.Printf("Invalid page number in client_slots callback: %s", pageStr)
			return
		}

		// Отправляем только будущие слоты для клиента
		shared.SendFutureSlotsForClient(h.ctx, h.b, h.userID, masterID, page, h.messageID)
		h.answerCallBackQuery("Обновлено", false)

	}
}
