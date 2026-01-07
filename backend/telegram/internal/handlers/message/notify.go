package message

import (
	"context"
	"fmt"
	"telegram-bot/internal/handlers/components"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// SendPlainMessage отправляет простое текстовое сообщение пользователю
func (h *Handler) SendPlainMessage(ctx context.Context, b *bot.Bot, userID int64, title, message string) {
	msg := fmt.Sprintf("%s🆕 Уведомление\n<b>%s</b><i>%s</i>\n", components.Header(), title, message)
	// text := title
	// if message != "" { if title != "" { text += "\n" } text += message }
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    userID,
		Text:      msg,
		ParseMode: models.ParseModeHTML,
	})
}

// SendRecordNotification отправляет уведомление о новой записи с кнопками действий (для мастера)
func (h *Handler) SendRecordNotification(ctx context.Context, b *bot.Bot, userID int64, recordID string, title, message string) {
	msg := fmt.Sprintf("%s🆕 Новая запись\n<b>%s</b>\n<i>%s</i>\n\nВыберите действие:", components.Header(), title, message)

	keyboard := &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
		{
			{Text: "✅ Подтвердить", CallbackData: fmt.Sprintf("record_action/confirm/%s", recordID)},
			{Text: "❌ Отклонить", CallbackData: fmt.Sprintf("record_action/reject/%s", recordID)},
		},
	}}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      userID,
		Text:        msg,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: keyboard,
	})
}

// SendRecordStatusNotification отправляет уведомление об изменении статуса записи (для клиента, без кнопок)
func (h *Handler) SendRecordStatusNotification(ctx context.Context, b *bot.Bot, userID int64, title, message string) {
	msg := fmt.Sprintf("%s🆕 Уведомление\n<b>%s</b>\n<i>%s</i>", components.Header(), title, message)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    userID,
		Text:      msg,
		ParseMode: models.ParseModeHTML,
	})
}

// SendAccountDeletionConfirmation отправляет запрос на подтверждение удаления аккаунта
func (h *Handler) SendAccountDeletionConfirmation(ctx context.Context, b *bot.Bot, userID int64, userUUID string) {
	msg := fmt.Sprintf(`%s⚠️ <b>Удаление аккаунта</b>

Вы запросили удаление своего аккаунта. Это действие <b>необратимо</b> и приведет к:

• Удалению всех ваших данных
• Удалению всех созданных слотов
• Удалению всех услуг
• Удалению всех записей
• Удалению уведомлений

<b>Вы уверены, что хотите продолжить?</b>

<i>Если вы случайно нажали кнопку удаления, просто проигнорируйте это сообщение.</i>`, components.Header())

	keyboard := &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
		{
			{Text: "❌ Отменить", CallbackData: "account_deletion/cancel"},
		},
		{
			{Text: "⚠️ ДА, УДАЛИТЬ АККАУНТ", CallbackData: fmt.Sprintf("account_deletion/confirm/%s", userUUID)},
		},
	}}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      userID,
		Text:        msg,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: keyboard,
	})
}
