package user

import (
	"sync"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository struct {
	db       *gorm.DB
	logger   *logrus.Logger
	mapToken *sync.Map
}

// NewUserRepository - конструктор Repository.
// Ответ: Возвращает ссылку на структуру Repository.
func NewRepository(db *gorm.DB, logger *logrus.Logger, mapToken *sync.Map) *Repository {
	return &Repository{
		db:       db,
		logger:   logger,
		mapToken: mapToken,
	}
}
