package user

import (
	"app/http/usecase/user"
	_ "fmt"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	service *user.Service
	logger  *logrus.Logger
}

func NewHandler(service *user.Service, logger *logrus.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}
