package callback

import (
	"context"
	"log"
	"strings"
	adapter "telegram-bot/internal/adapter/backendapi"
	mybot "telegram-bot/internal/bot"
	"telegram-bot/internal/config"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"
)

// UniversalHandler обработчик всех Inline/Reply keyboard нажатий
func UniversalHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	cfg := config.Load()
	client := adapter.New(cfg.BackendBaseURL, logrus.New())
	handler := NewHandler(*client, cfg, ctx, b, update)
	if update == nil {
		log.Printf("UniversalHandler: " + "error, update is nil")
		return
	}

	// rate limit
	if update.CallbackQuery != nil {
		userID := update.CallbackQuery.From.ID
		if !mybot.CanExecuteCallback(userID) {
			answerCallBackQuery(ctx, b, update, "⏳ Слишком быстро! Подождите немного.", true)
			return
		}
	}

	// Обработка событий нажатия на ReplyKeyboard «Поделиться контактом»
	if update.Message != nil && update.Message.Contact != nil {
		handler.ContactHandler()
	}

	// Обработка событий нажатия на InlineKeyboard
	if update.CallbackQuery != nil {
		callbackQuery := update.CallbackQuery
		callbackData := callbackQuery.Data

		var userID int64
		if callbackQuery.From.ID != 0 {
			userID = callbackQuery.From.ID
		} else {
			log.Println("CallbackQuery.From.ID is zero")
			return
		}

		// Получаем ID сообщения, на котором была нажата кнопка
		var messageID int
		if callbackQuery.Message.Message != nil {
			messageID = callbackQuery.Message.Message.ID
		}
		callbackHandler := NewCallBackHandler(*handler, messageID, userID, callbackData)

		// Событие Подтвердить вход
		if callbackData == "ConfirmLogin" {
			callbackHandler.TryConfirmLogin()
		}
		// Выбор таймзоны: tz/{?}
		if strings.HasPrefix(callbackData, "tz/") {
			callbackHandler.SelectTimezone()
		}
		// Обработка навигации по датам и страницам слотов
		if strings.HasPrefix(callbackData, "date/") {
			callbackHandler.DateMove()
		}
		// Обработка навигации по датам для клиентского просмотра (только будущие)
		if strings.HasPrefix(callbackData, "client_date/") {
			callbackHandler.DateMoveClient()
		}
		// Обработка выбора конкретного слота
		if strings.HasPrefix(callbackData, "slot/") {
			callbackHandler.SlotMove()
		}
		// Обработка создания записи на слот
		if strings.HasPrefix(callbackData, "book/") {
			callbackHandler.BookMove()
		}
		// Обработка кнопки "Назад к слотам"
		if strings.HasPrefix(callbackData, "back_to_slots/") {
			callbackHandler.BackToSlots()
		}
		// Обработка неактивных кнопок
		if callbackData == "noop" {
			b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: callbackQuery.ID,
				Text:            "Текущая страница",
			})
		}
		// Пагинация по записям пользователя: records/{status}/{page}
		if strings.HasPrefix(callbackData, "records/") {
			callbackHandler.Records()
		}
		// Пулы записей по времени: records_time/{future|past|chooser}/{status}/{page}
		if strings.HasPrefix(callbackData, "records_time/") {
			callbackHandler.RecordsTime()
		}
		// Обработка слотов по времени: slots_time/{future|past|chooser}/{masterID}/{page}
		if strings.HasPrefix(callbackData, "slots_time/") {
			callbackHandler.SlotsTime()
		}
		// Обработка клиентских слотов: client_slots/{masterID}/{page}
		if strings.HasPrefix(callbackData, "client_slots/") {
			callbackHandler.ClientSlots()
		}
		// Обработка всех записей по времени: all_records_time/{future|past|chooser}/{status}/{page}
		if strings.HasPrefix(callbackData, "all_records_time/") {
			callbackHandler.AllRecordsTime()
		}
		// Обработка действий с записями: record_action/{confirm|reject}/{recordID}
		if strings.HasPrefix(callbackData, "record_action/") {
			callbackHandler.RecordAction()
		}
		// Обработка удаления аккаунта: account_deletion/{cancel|confirm}/{userUUID}
		if strings.HasPrefix(callbackData, "account_deletion/") {
			callbackHandler.AccountDeletion()
		}
		// Обработка пагинации предстоящих записей мастера: upcoming_page/{page}/{telegramID}
		if strings.HasPrefix(callbackData, "upcoming_page/") {
			callbackHandler.UpcomingPage()
		}
	}
}
