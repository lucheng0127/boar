package agent

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime/debug"
	"time"

	api "github.com/osrg/gobgp/v3/api"
	"github.com/sirupsen/logrus"
)

type PathInfo struct {
	RT          string
	RD          string
	IP          net.IP
	Mac         net.HardwareAddr
	Age         int64
	Withdrawal  bool
	Neighbor    net.IP
	IsLocal     bool
	Nexthop     net.IP
	AsPath      string
	Communities []uint32
}

type AgentServer struct {
	host   string
	logger *logrus.Logger
	pathCh chan *PathInfo
}

func (p *PathInfo) Detail() string {
	t := time.Unix(p.Age, 0)
	tStr := t.UTC().Format(time.RFC3339)
	return fmt.Sprintf(
		"age [%s] rt [%s] rd [%s] ip [%s] mac [%s] withdrawl [%t] neighbor [%s] local [%t] nexthop [%s] as path [%s] communities %+v",
		tStr, p.RT, p.RD, p.Mac.String(), p.IP.String(),
		p.Withdrawal, p.Neighbor.String(), p.IsLocal, p.Nexthop.String(), p.AsPath, p.Communities,
	)
}

func NewServer(host string, logger *logrus.Logger) *AgentServer {
	return &AgentServer{
		host:   host,
		logger: logger,
		pathCh: make(chan *PathInfo, 16),
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
	Monitor(recver, s)
}

func (s *AgentServer) handleEVPNMsg() {
	for {
		path := <-s.pathCh
		s.logger.WithFields(logrus.Fields{
			"Topic": "BGP MSG",
		}).Debugf("Receive EVPN MSG, pathinfo %s", path.Detail())

		if path.IsLocal {
			s.logger.WithFields(logrus.Fields{
				"Topic": "BGP MSG",
			}).Debugf("Skip handle local bgp msg %s", path.Detail())
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
