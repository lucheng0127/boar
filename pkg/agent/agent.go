package agent

import (
	"os"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

type AgentServer struct {
	logger *logrus.Logger
}

func NewServer(logger *logrus.Logger) *AgentServer {
	return &AgentServer{
		logger: logger,
	}
}

func (s *AgentServer) Serve() {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Errorf("%s\n%s", r, string(debug.Stack()))
			os.Exit(1)
		}
	}()

	s.logger.WithFields(logrus.Fields{
		"Topic": "Agent",
	}).Info("Agent Start ...")

	// Monitor rib

	// Handle rib
}
