package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type recordNotifyRequest struct {
	RecordID   uint   `json:"record_id"`
	TelegramID int64  `json:"telegram_id"`
	Title      string `json:"title"`
	Message    string `json:"message"`
}

type recordStatusNotifyRequest struct {
	TelegramID int64  `json:"telegram_id"`
	Title      string `json:"title"`
	Message    string `json:"message"`
}

type accountDeletionRequest struct {
	UserID     string `json:"user_id"`
	TelegramID int64  `json:"telegram_id"`
}

// NotifyRecord принимает POST-запрос и отправляет сообщение пользователю в телеграм
func (h *HttpClient) NotifyRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	var req recordNotifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректное тело запроса", http.StatusBadRequest)
		return
	}
	if req.TelegramID == 0 {
		http.Error(w, "telegram_id обязателен", http.StatusBadRequest)
		return
	}
	log.Printf("NotifyRecord: to=%d record_id=%d title=%q", req.TelegramID, req.RecordID, req.Title)

	// Для уведомлений о записях всегда отправляем с кнопками действий
	// Если record_id не передан, используем заглушку
	recordID := strconv.Itoa(int(req.RecordID))
	if recordID == "0" {
		recordID = "unknown"
	}

	h.messageHandler.SendRecordNotification(r.Context(), h.bot, req.TelegramID, recordID, req.Title, req.Message)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf("ok")))
}

// NotifyRecordStatus принимает POST-запрос и отправляет уведомление о статусе записи (без кнопок)
func (h *HttpClient) NotifyRecordStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	var req recordStatusNotifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректное тело запроса", http.StatusBadRequest)
		return
	}
	if req.TelegramID == 0 {
		http.Error(w, "telegram_id обязателен", http.StatusBadRequest)
		return
	}
	log.Printf("NotifyRecordStatus: to=%d title=%q", req.TelegramID, req.Title)

	h.messageHandler.SendRecordStatusNotification(r.Context(), h.bot, req.TelegramID, req.Title, req.Message)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf("ok")))
}

// NotifyAccountDeletion принимает POST-запрос и отправляет запрос на подтверждение удаления аккаунта
func (h *HttpClient) NotifyAccountDeletion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	var req accountDeletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректное тело запроса", http.StatusBadRequest)
		return
	}
	if req.TelegramID == 0 || req.UserID == "" {
		http.Error(w, "telegram_id и user_id обязательны", http.StatusBadRequest)
		return
	}
	log.Printf("NotifyAccountDeletion: to=%d user_id=%s", req.TelegramID, req.UserID)

	h.messageHandler.SendAccountDeletionConfirmation(r.Context(), h.bot, req.TelegramID, req.UserID)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf("ok")))
}
