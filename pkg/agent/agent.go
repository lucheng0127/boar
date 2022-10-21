package agent

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	api "github.com/osrg/gobgp/v3/api"
	"github.com/osrg/gobgp/v3/pkg/apiutil"
	"github.com/sirupsen/logrus"

	"github.com/lucheng0127/boar/pkg/utils"
)

type AgentServer struct {
	host   string
	logger *logrus.Logger
	pathCh chan *apiutil.Path
}

func NewServer(host string, logger *logrus.Logger) *AgentServer {
	return &AgentServer{
		host:   host,
		logger: logger,
		pathCh: make(chan *apiutil.Path, 64),
	}
}

func getClient(host string, interval int, logger *logrus.Logger) api.GobgpApiClient {
	ctx := context.Background()
	if client, cancel, err := NewClient(ctx, host); err == nil {
		logger.WithFields(logrus.Fields{
			"Topic": "Agent",
		}).Info("Connect to gobgp server")
		return client
	} else {
		logger.WithFields(logrus.Fields{
			"Topic": "Agent",
		}).Errorf("Failed to connect to gobgp server %s, retry afer %d seconds", host, interval)
		cancel()
		time.Sleep(time.Duration(interval) * time.Second)
		interval *= 2
		return getClient(host, interval, logger)
	}
}

func (s *AgentServer) monitorRibs() {
	ctx := context.Background()
	c := getClient(s.host, 1, s.logger)
	recver, err := c.WatchEvent(ctx, &api.WatchEventRequest{
		Table: &api.WatchEventRequest_Table{
			Filters: []*api.WatchEventRequest_Table_Filter{
				{
					Type: api.WatchEventRequest_Table_Filter_BEST,
					Init: true,
				},
			},
		},
	})
	if err != nil {
		panic(fmt.Sprintf("Client get watch event failed: %s", err.Error()))
	}
	Monitor(recver, s.pathCh)
}

func (s *AgentServer) handleEVPNMsg() {
	for {
		path := <-s.pathCh
		s.logger.WithFields(logrus.Fields{
			"Topic": "BGP MSG",
		}).Debugf("Receive EVPN MSG, pathinfo [%s]", utils.PathDetail(path))

		nexthop := utils.GetNextHopFromPathAttributes(path.Attrs)
		if shouldSkip, err := utils.IsLocalIP(nexthop.String()); err != nil {
			s.logger.WithFields(logrus.Fields{
				"Topic": "Config",
			}).Errorf("Failed to get local IP address")
			panic(err)
		} else if shouldSkip {
			s.logger.WithFields(logrus.Fields{
				"Topic": "BGP MSG",
			}).Debugf("Skip handle local bgp msg [%s]", utils.PathDetail(path))
			continue
		}

		// Handle MSG
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

	// Initial
	go s.monitorRibs()

	// Monitor rib
	go s.handleEVPNMsg()
}
