package server

import (
	"github.com/lucheng0127/boar/pkg/agent"
	"github.com/lucheng0127/boar/pkg/api"

	"github.com/sirupsen/logrus"
)

type Server interface {
	Serve()
}

type BoarServer struct {
	api   *api.APIServer
	agent *agent.AgentServer
}

func NewServer(port int, host string, logger *logrus.Logger) *BoarServer {
	return &BoarServer{
		api:   api.NewServer(port, logger),
		agent: agent.NewServer(host, logger),
	}
}

func (s *BoarServer) Serve() {
	go s.api.Serve()
	go s.agent.Serve()
}
