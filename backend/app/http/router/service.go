package router

import (
	serviceCtrl "app/http/controller/service"
	serviceRepo "app/http/repository/service"
	userRepo "app/http/repository/user"
	serviceServ "app/http/usecase/service"
	"sync"
)

func (s *Client) GetServiceHandler(tokenMap *sync.Map) *serviceCtrl.Handler {
	Repo := serviceRepo.NewRepository(s.gormDB, s.logger)
	UserRepo := userRepo.NewRepository(s.gormDB, s.logger, tokenMap)
	Serv := serviceServ.NewService(Repo, UserRepo, s.logger)
	Ctrl := serviceCtrl.NewHandler(Serv, s.logger)
	return Ctrl
}
