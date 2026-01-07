package service

import (
	"app/http/usecase/service"
	_ "fmt"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	service *service.Service
	logger  *logrus.Logger
}

func NewHandler(service *service.Service, logger *logrus.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}
