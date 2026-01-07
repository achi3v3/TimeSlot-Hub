package router

import (
	recordCtrl "app/http/controller/record"
	notifyRepo "app/http/repository/notification"
	recordRepo "app/http/repository/record"
	notifyServ "app/http/usecase/notification"
	recordServ "app/http/usecase/record"
)

func (s *Client) GetRecordHandler() *recordCtrl.Handler {
	notificationRepo := notifyRepo.NewRepository(s.gormDB, s.logger)
	notificationService := notifyServ.NewService(notificationRepo, s.logger)
	Repo := recordRepo.NewRepository(s.gormDB, s.logger)
	Serv := recordServ.NewService(Repo, notificationService, s.logger)
	Ctrl := recordCtrl.NewHandler(Serv, s.logger)
	return Ctrl
}
