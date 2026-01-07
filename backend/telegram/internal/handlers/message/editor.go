package message

import (
	"context"
	"log"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// MessageEditor предоставляет методы для безопасного редактирования сообщений
type MessageEditor struct {
	stateManager *MessageStateManager
}

// NewMessageEditor создает новый экземпляр редактора сообщений
func NewMessageEditor() *MessageEditor {
	return &MessageEditor{
		stateManager: GetStateManager(),
	}
}

// SafeEditMessage безопасно редактирует сообщение с обработкой ошибок
func (e *MessageEditor) SafeEditMessage(ctx context.Context, b *bot.Bot, chatID int64, messageID int, text string, replyMarkup *models.InlineKeyboardMarkup) error {
	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   messageID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: replyMarkup,
	})

	if err != nil {
		// Если контент не изменился — не считаем это ошибкой и не шлём новое сообщение
		if strings.Contains(strings.ToLower(err.Error()), "message is not modified") {
			log.Printf("Skip edit for message %d in chat %d: not modified", messageID, chatID)
			return nil
		}
		log.Printf("Failed to edit message %d in chat %d: %v", messageID, chatID, err)
		return err
	}

	return nil
}

// SendOrEditMessage отправляет новое сообщение или редактирует существующее
func (e *MessageEditor) SendOrEditMessage(ctx context.Context, b *bot.Bot, chatID, userID int64, messageType string, text string, replyMarkup *models.InlineKeyboardMarkup, context map[string]interface{}) error {
	// Проверяем, есть ли активное сообщение этого типа
	if state, exists := e.stateManager.GetMessageState(chatID, userID, messageType); exists {
		log.Printf("Found existing message state for %s: chatID=%d, messageID=%d", messageType, state.ChatID, state.MessageID)

		// Пытаемся отредактировать существующее сообщение
		err := e.SafeEditMessage(ctx, b, state.ChatID, state.MessageID, text, replyMarkup)
		if err == nil {
			log.Printf("Successfully edited message %d in chat %d for type %s", state.MessageID, state.ChatID, messageType)
			// Обновляем контекст состояния
			e.stateManager.UpdateMessageState(chatID, userID, messageType, context)
			return nil
		}

		log.Printf("Failed to edit message %d in chat %d for type %s: %v", state.MessageID, state.ChatID, messageType, err)
		// Если редактирование не удалось, удаляем старое состояние
		e.stateManager.RemoveMessageState(chatID, userID, messageType)
	}

	// Отправляем новое сообщение
	log.Printf("Sending new message to user %d, chat %d for type %s", userID, chatID, messageType)
	sentMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: replyMarkup,
	})

	if err != nil {
		log.Printf("Failed to send message to user %d: %v", userID, err)
		return err
	}

	log.Printf("Successfully sent new message %d to chat %d for user %d, type %s", sentMsg.ID, chatID, userID, messageType)
	// Сохраняем состояние нового сообщения
	e.stateManager.SetMessageState(chatID, userID, sentMsg.ID, messageType, context)

	return nil
}

// EditSlotsPagination редактирует сообщение с пагинацией слотов
func (e *MessageEditor) EditSlotsPagination(ctx context.Context, b *bot.Bot, userID, masterID int64, targetDate string, page int, text string, replyMarkup *models.InlineKeyboardMarkup) error {
	context := map[string]interface{}{
		"masterID":   masterID,
		"targetDate": targetDate,
		"page":       page,
	}

	return e.SendOrEditMessage(ctx, b, userID, userID, "slots_pagination", text, replyMarkup, context)
}

// EditSlotDetails редактирует сообщение с деталями слота
func (e *MessageEditor) EditSlotDetails(ctx context.Context, b *bot.Bot, userID int64, slotID uint, text string, replyMarkup *models.InlineKeyboardMarkup) error {
	context := map[string]interface{}{
		"slotID": slotID,
	}

	// ВАЖНО: редактируем то же сообщение, что и пагинация, чтобы не плодить новые
	return e.SendOrEditMessage(ctx, b, userID, userID, "slots_pagination", text, replyMarkup, context)
}

// EditUserRecords редактирует/отправляет сообщение со списком записей пользователя
func (e *MessageEditor) EditUserRecords(
	ctx context.Context,
	b *bot.Bot,
	userID int64,
	status string,
	page int,
	text string,
	replyMarkup *models.InlineKeyboardMarkup,
) error {
	context := map[string]interface{}{
		"status": status,
		"page":   page,
	}
	return e.SendOrEditMessage(ctx, b, userID, userID, "user_records", text, replyMarkup, context)
}

// EditMessage редактирует сообщение по типу
func (e *MessageEditor) EditMessage(ctx context.Context, b *bot.Bot, userID int64, messageType string, text string, replyMarkup *models.InlineKeyboardMarkup) error {
	// Проверяем, есть ли активное сообщение этого типа
	if state, exists := e.stateManager.GetMessageState(userID, userID, messageType); exists {
		log.Printf("Found existing message state for %s: chatID=%d, messageID=%d", messageType, state.ChatID, state.MessageID)

		// Пытаемся отредактировать существующее сообщение
		err := e.SafeEditMessage(ctx, b, state.ChatID, state.MessageID, text, replyMarkup)
		if err == nil {
			log.Printf("Successfully edited message %d in chat %d for type %s", state.MessageID, state.ChatID, messageType)
			return nil
		}

		log.Printf("Failed to edit message %d in chat %d for type %s: %v", state.MessageID, state.ChatID, messageType, err)
		// Если редактирование не удалось, удаляем старое состояние
		e.stateManager.RemoveMessageState(userID, userID, messageType)
	}

	// Отправляем новое сообщение
	log.Printf("Sending new message to user %d for type %s", userID, messageType)
	sentMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      userID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: replyMarkup,
	})

	if err != nil {
		log.Printf("Failed to send message to user %d: %v", userID, err)
		return err
	}

	log.Printf("Successfully sent new message %d to chat %d for user %d, type %s", sentMsg.ID, userID, userID, messageType)
	// Сохраняем состояние нового сообщения
	e.stateManager.SetMessageState(userID, userID, sentMsg.ID, messageType, map[string]interface{}{})

	return nil
}

// EditSpecificMessage редактирует конкретное сообщение по ID
func (e *MessageEditor) EditSpecificMessage(ctx context.Context, b *bot.Bot, chatID int64, messageID int, text string, replyMarkup *models.InlineKeyboardMarkup) error {
	// Пытаемся отредактировать конкретное сообщение
	err := e.SafeEditMessage(ctx, b, chatID, messageID, text, replyMarkup)
	if err == nil {
		log.Printf("Successfully edited message %d in chat %d", messageID, chatID)
		return nil
	}

	log.Printf("Failed to edit message %d in chat %d: %v", messageID, chatID, err)
	// Если редактирование не удалось, отправляем новое сообщение
	log.Printf("Sending new message to chat %d", chatID)
	sentMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: replyMarkup,
	})

	if err != nil {
		log.Printf("Failed to send message to chat %d: %v", chatID, err)
		return err
	}

	log.Printf("Successfully sent new message %d to chat %d", sentMsg.ID, chatID)
	return nil
}

// RemoveMessageState удаляет состояние сообщения (например, после успешного бронирования)
func (e *MessageEditor) RemoveMessageState(userID int64, messageType string) {
	e.stateManager.RemoveMessageState(userID, userID, messageType)
}
