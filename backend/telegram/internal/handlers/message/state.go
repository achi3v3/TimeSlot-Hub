package message

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MessageState хранит информацию о состоянии сообщения
type MessageState struct {
	MessageID   int
	ChatID      int64
	UserID      int64
	MessageType string                 // "slots_pagination", "slot_details", "booking_confirmation"
	Context     map[string]interface{} // дополнительные данные контекста
	CreatedAt   time.Time
	LastUpdated time.Time
}

// MessageStateManager управляет состояниями сообщений
type MessageStateManager struct {
	states        map[string]*MessageState // ключ: chatID_userID_messageType
	mutex         sync.RWMutex
	cleanupCtx    context.Context
	cleanupCancel context.CancelFunc
	name          string
}

var globalStateManager = &MessageStateManager{
	states: make(map[string]*MessageState),
}

// GetStateManager возвращает глобальный менеджер состояний
func GetStateManager() *MessageStateManager {
	if globalStateManager.name == "" {
		globalStateManager.name = "message-state-manager"
	}
	return globalStateManager
}

// generateKey создает уникальный ключ для состояния
func (m *MessageStateManager) generateKey(chatID, userID int64, messageType string) string {
	return fmt.Sprintf("%d_%d_%s", chatID, userID, messageType)
}

// SetMessageState сохраняет состояние сообщения
func (m *MessageStateManager) SetMessageState(chatID, userID int64, messageID int, messageType string, context map[string]interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := m.generateKey(chatID, userID, messageType)
	now := time.Now()

	m.states[key] = &MessageState{
		MessageID:   messageID,
		ChatID:      chatID,
		UserID:      userID,
		MessageType: messageType,
		Context:     context,
		CreatedAt:   now,
		LastUpdated: now,
	}
}

// GetMessageState получает состояние сообщения
func (m *MessageStateManager) GetMessageState(chatID, userID int64, messageType string) (*MessageState, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	key := m.generateKey(chatID, userID, messageType)
	state, exists := m.states[key]
	return state, exists
}

// UpdateMessageState обновляет существующее состояние
func (m *MessageStateManager) UpdateMessageState(chatID, userID int64, messageType string, context map[string]interface{}) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := m.generateKey(chatID, userID, messageType)
	if state, exists := m.states[key]; exists {
		state.Context = context
		state.LastUpdated = time.Now()
		return true
	}
	return false
}

// RemoveMessageState удаляет состояние сообщения
func (m *MessageStateManager) RemoveMessageState(chatID, userID int64, messageType string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := m.generateKey(chatID, userID, messageType)
	delete(m.states, key)
}

// CleanupOldStates удаляет старые состояния (старше указанного времени)
func (m *MessageStateManager) CleanupOldStates(maxAge time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	for key, state := range m.states {
		if now.Sub(state.LastUpdated) > maxAge {
			delete(m.states, key)
		}
	}
}

// StartCleanupRoutine запускает горутину для периодической очистки старых состояний
func (m *MessageStateManager) StartCleanupRoutine() {
	m.cleanupCtx, m.cleanupCancel = context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(30 * time.Minute) // очистка каждые 30 минут
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.CleanupOldStates(24 * time.Hour) // удаляем состояния старше 24 часов
			case <-m.cleanupCtx.Done():
				return
			}
		}
	}()
}

func (m *MessageStateManager) Close(ctx context.Context) error {
	if m.cleanupCancel != nil {
		m.cleanupCancel()
	}
	return nil
}
func (m *MessageStateManager) Name() string { return m.name }
