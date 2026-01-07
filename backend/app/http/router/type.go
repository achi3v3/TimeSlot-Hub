package router

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type HttpServer struct {
	server *http.Server
	name   string
}

func NewHTTPServer(server *http.Server, name string) *HttpServer {
	return &HttpServer{
		server: server,
		name:   name,
	}
}
func (s *HttpServer) Name() string                       { return s.name }
func (s *HttpServer) Shutdown(ctx context.Context) error { return s.server.Shutdown(ctx) }

type Client struct {
	DataBase
	router     *gin.Engine
	logger     *logrus.Logger
	httpServer *HttpServer
}

type DataBase struct {
	gormDB *gorm.DB
}

// DB возвращает gorm.DB
func (s *Client) DB() *gorm.DB {
	return s.gormDB
}

func NewClient(db *gorm.DB, logger *logrus.Logger, httpServer *HttpServer, r *gin.Engine) *Client {
	return &Client{
		router:     r,
		DataBase:   DataBase{gormDB: db},
		logger:     logger,
		httpServer: httpServer,
	}
}
