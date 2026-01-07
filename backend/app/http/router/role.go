package router

import (
	roleCtrl "app/http/controller/role"
)

func (s *Client) GetRoleHandler() *roleCtrl.Handler {
	Ctrl := roleCtrl.NewHandler(s.gormDB)
	return Ctrl
}
