package router

import (
	slotCtrl "app/http/controller/slot"
	notifyRepo "app/http/repository/notification"
	recordRepo "app/http/repository/record"
	slotRepo "app/http/repository/slot"
	notifyServ "app/http/usecase/notification"
	slotServ "app/http/usecase/slot"
)

func (s *Client) GetSlotHandler() *slotCtrl.Handler {
	Repo := slotRepo.NewRepository(s.gormDB, s.logger)
	// wire notification service and record repo for deletion notifications
	nRepo := notifyRepo.NewRepository(s.gormDB, s.logger)
	nServ := notifyServ.NewService(nRepo, s.logger)
	rRepo := recordRepo.NewRepository(s.gormDB, s.logger)

	Serv := slotServ.NewService(Repo, s.logger).WithNotification(nServ).WithRecordRepository(rRepo)
	Ctrl := slotCtrl.NewHandler(Serv, s.logger)
	return Ctrl
}
