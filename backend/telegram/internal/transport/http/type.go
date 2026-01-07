package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"telegram-bot/internal/handlers/message"
	"time"

	"github.com/go-telegram/bot"
	"github.com/sirupsen/logrus"
)

const (
	notifyLink = "/notify"
)

type HttpClient struct {
	bot            *bot.Bot
	messageHandler *message.Handler
	logger         *logrus.Logger
	name           string
	server         *http.Server
}

func NewHttpClient(name string, server *http.Server, bot *bot.Bot, messageHandler *message.Handler, logger *logrus.Logger) *HttpClient {
	return &HttpClient{
		bot:            bot,
		messageHandler: messageHandler,
		logger:         logger,
		name:           name,
		server:         server,
	}
}

func (h *HttpClient) Name() string                       { return h.name }
func (h *HttpClient) Shutdown(ctx context.Context) error { return h.server.Shutdown(ctx) }

// StartHttp метод Server. Запускает HTTP сервер
func (h *HttpClient) Run() error {
	mux, ok := h.server.Handler.(*http.ServeMux)
	if !ok {
		mux = http.NewServeMux()
		h.server.Handler = mux
	}

	shared := os.Getenv("TELEGRAM_HTTP_SECRET")
	checkSecret := func(r *http.Request) bool {
		if shared == "" {
			return true
		}
		return r.Header.Get("X-Internal-Token") == shared
	}
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	mux.HandleFunc(notifyLink+"-login/", func(w http.ResponseWriter, r *http.Request) {
		if !checkSecret(r) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("forbidden"))
			return
		}
		h.NotifyLogin(w, r)
	})
	mux.HandleFunc(notifyLink+"-record", func(w http.ResponseWriter, r *http.Request) {
		if !checkSecret(r) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("forbidden"))
			return
		}
		h.NotifyRecord(w, r)
	})
	mux.HandleFunc(notifyLink+"-record-status", func(w http.ResponseWriter, r *http.Request) {
		if !checkSecret(r) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("forbidden"))
			return
		}
		h.NotifyRecordStatus(w, r)
	})
	mux.HandleFunc(notifyLink+"-account-deletion", func(w http.ResponseWriter, r *http.Request) {
		if !checkSecret(r) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("forbidden"))
			return
		}
		h.NotifyAccountDeletion(w, r)
	})

	h.logger.Infof("HTTP сервер для нотификаций запущен на :8091%v-login/{telegram_id}, POST %v-record, POST %v-record-status, POST %v-account-deletion", notifyLink, notifyLink, notifyLink, notifyLink)
	if err := h.server.ListenAndServe(); err != nil {
		return fmt.Errorf("Ошибка запуска сервера: %v", err)
	}
	return nil
}
