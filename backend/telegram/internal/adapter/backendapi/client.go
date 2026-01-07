package backendapi

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Client struct {
	baseURL string
	http    *http.Client
	logger  *logrus.Logger
}

func New(baseURL string, logger *logrus.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
		logger:  logger,
	}
}
