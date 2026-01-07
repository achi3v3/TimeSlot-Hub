package notification

import (
	"app/http/usecase/notification"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	service *notification.Service
	logger  *logrus.Logger
}

func NewHandler(service *notification.Service, logger *logrus.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}
