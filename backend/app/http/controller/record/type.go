package record

import (
	"app/http/usecase/record"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	service *record.Service
	logger  *logrus.Logger
}

func NewHandler(service *record.Service, logger *logrus.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}
