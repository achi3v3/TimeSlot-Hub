package service

import (
	"app/http/repository/service"
	"app/http/repository/user"

	"github.com/sirupsen/logrus"
)

type Service struct {
	repo     *service.Repository
	userRepo *user.Repository
	logger   *logrus.Logger
}

func NewService(repo *service.Repository, userRepo *user.Repository, logger *logrus.Logger) *Service {
	return &Service{
		repo:     repo,
		userRepo: userRepo,
		logger:   logger,
	}
}
