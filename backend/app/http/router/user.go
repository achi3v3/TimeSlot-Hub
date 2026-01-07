package router

import (
	userCtrl "app/http/controller/user"
	userRepo "app/http/repository/user"
	userServ "app/http/usecase/user"
	"sync"
)

var tokenMap = &sync.Map{}

func (s *Client) GetUserHandler() *userCtrl.Handler {
	Repo := userRepo.NewRepository(s.gormDB, s.logger, tokenMap)
	Serv := userServ.NewService(Repo, s.logger)
	Ctrl := userCtrl.NewHandler(Serv, s.logger)
	return Ctrl
}
