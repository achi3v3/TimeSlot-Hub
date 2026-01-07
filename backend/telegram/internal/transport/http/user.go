package http

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// NotifyLogin метод Server. Обрабатывает GET-запрос на подтверждение входа пользователя
func (h *HttpClient) NotifyLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	telegramIDStr := r.URL.Path[len(notifyLink+"-login")+1:] // .../notify-login/123 -> 123
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Некорректный telegram_id", http.StatusBadRequest)
		return
	}
	log.Printf("Received TelegramID: %d\n", telegramID)
	// optional ip/location from query
	ip := r.URL.Query().Get("ip")
	loc := r.URL.Query().Get("loc")
	h.messageHandler.SendLoginMessage(r.Context(), h.bot, telegramID, ip, loc)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Уведомление отправлено для пользователя %d", telegramID)))
}
