package api

import (
	"os"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

type APIServer struct {
	port   int
	logger *logrus.Logger
}

func NewServer(port int, logger *logrus.Logger) *APIServer {
	return &APIServer{
		port:   port,
		logger: logger,
	}
}

func (s *APIServer) Serve() {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Errorf("%s\n%s", r, string(debug.Stack()))
			os.Exit(1)
		}
	}()

	s.logger.WithFields(logrus.Fields{
		"Topic": "API",
	}).Infof("API server run on port %d ...", s.port)
}
