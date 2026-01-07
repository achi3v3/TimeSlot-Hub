package router

import (
	notifyCtrl "app/http/controller/notification"
	notifyRepo "app/http/repository/notification"
	notifyServ "app/http/usecase/notification"
)

func (s *Client) GetNotificationHandler() *notifyCtrl.Handler {
	Repo := notifyRepo.NewRepository(s.gormDB, s.logger)
	Serv := notifyServ.NewService(Repo, s.logger)
	Ctrl := notifyCtrl.NewHandler(Serv, s.logger)
	return Ctrl
}
