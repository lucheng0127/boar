package agent

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime/debug"
	"time"

	"github.com/lucheng0127/boar/pkg/utils"
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
	IsBUM       bool
}

type AgentServer struct {
	host   string
	logger *logrus.Logger
	pathCh chan *PathInfo
	vniMap map[int][]int
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
		vniMap: make(map[int][]int),
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

func (s *AgentServer) handleBUM(p *PathInfo) {
	// Get L3VNI from rt mark it, then we should handle BGP MSG that with same rt
	// Use vpc vni as rt, there can be more then one vxnet under a vpc
	s.logger.WithFields(logrus.Fields{
		"Topic": "BUM",
	}).Debug(p.Detail())
	_, rt, err := utils.ParseVni(p.RT)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"Topic": "BUM",
		}).Errorf("Parse vni failed %s", err.Error())
		return
	}

	vtep, err := utils.GetVtepByVNI(rt)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"Topic": "BUM",
		}).Errorf("Get vtep by vni [%d] failed", rt)
		return
	}
	if len(vtep) == 0 {
		s.logger.WithFields(logrus.Fields{
			"Topic": "BUM",
		}).Errorf("No vxlan found with vni [%d], invalid bgp msg %s", rt, p.Detail())
		return
	}
	s.vniMap[rt] = append(s.vniMap[rt], rt)
}

func (s *AgentServer) handleDVR() {
	fmt.Println("Not Implement")
}

func (s *AgentServer) handleVM() {
	fmt.Println("Not Implement")
}

func (s *AgentServer) handleMacadv(p *PathInfo) {
	s.logger.WithFields(logrus.Fields{
		"Topic": "MACADV",
	}).Debug(p.Detail())
	_, rt, err := utils.ParseVni(p.RT)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"Topic": "MACADV",
		}).Errorf("Parse vni failed %s", err.Error())
		return
	}
	if _, ok := s.vniMap[rt]; !ok {
		s.logger.WithFields(logrus.Fields{
			"Topic": "MACADV",
			"RT":    rt,
		}).Debugf("Skip handle bgp msg %s", p.Detail())
	} else {
		vniType, rd, err := utils.ParseVni(p.RD)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"Topic": "MACADV",
			}).Errorf("Parse vni failed %s", err.Error())
			return
		}
		s.vniMap[rt] = append(s.vniMap[rt], rd)
		switch vniType {
		case utils.MSG_TYPE_DVR:
			s.handleDVR()
		case utils.MSG_TYPE_VM:
			s.handleVM()
		default:
			s.logger.WithFields(logrus.Fields{
				"Topic": "MACADV",
			}).Warnf("Unspported msg type %s", p.Detail())
		}
	}
}

func (s *AgentServer) handleEVPNMsg() {
	defer func() {
		if r := recover(); r != nil {
			s.logger.WithFields(logrus.Fields{
				"Topic": "Handle BGP MSG",
			}).Error("Failed to handle MSG ", r)
		}
	}()

	for {
		p := <-s.pathCh
		s.logger.WithFields(logrus.Fields{
			"Topic": "BGP MSG",
		}).Debug(p.Detail())

		if p.IsLocal {
			s.logger.WithFields(logrus.Fields{
				"Topic": "Local BGP",
			}).Debug(p.Detail())
			continue
		}

		// Handle MSG
		if p.IsBUM {
			s.handleBUM(p)
		} else {
			s.handleMacadv(p)
		}
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
