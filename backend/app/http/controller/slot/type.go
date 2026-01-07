package slot

import (
	"app/http/usecase/slot"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	service *slot.Service
	logger  *logrus.Logger
}

func NewHandler(service *slot.Service, logger *logrus.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}
