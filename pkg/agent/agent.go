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

type vxlanInfo struct {
	vni        int // Remote vxlan peer vni
	ip         net.IP
	mac        net.HardwareAddr
	nexthop    net.IP
	vtep       string // Local vxlan vtep
	netnsId    int    // Local dvr netns id
	withdrawal bool
	dvrIface   string
	netLen     int
	subnetLen  int
}

type AgentServer struct {
	host    string
	logger  *logrus.Logger
	pathCh  chan *PathInfo
	vniMap  map[int][]int
	msgBook map[string]int64 // Mac-ip as key, age as value
}

func (info *vxlanInfo) Detail() string {
	return fmt.Sprintf(
		"vni [%d] ip [%s] mac [%s] nexthop [%s] vtep [%s] netnsId [%d] withdrawal [%t] net len [%d] subnet len [%d]",
		info.vni, info.ip.String(), info.mac.String(), info.nexthop.String(), info.vtep, info.netnsId, info.withdrawal, info.netLen, info.subnetLen,
	)
}

func (p *PathInfo) Detail() string {
	t := time.Unix(p.Age, 0)
	tStr := t.Format(utils.CLS_FORMAT)
	return fmt.Sprintf(
		"age [%s] rt [%s] rd [%s] ip [%s] mac [%s] withdrawl [%t] neighbor [%s] local [%t] nexthop [%s] as path [%s] communities %+v",
		tStr, p.RT, p.RD, p.Mac.String(), p.IP.String(),
		p.Withdrawal, p.Neighbor.String(), p.IsLocal, p.Nexthop.String(), p.AsPath, p.Communities,
	)
}

func (p *PathInfo) Identification() string {
	return fmt.Sprintf("%s-%s", p.Mac.String(), p.IP.String())
}

func NewServer(host string, logger *logrus.Logger) *AgentServer {
	return &AgentServer{
		host:    host,
		logger:  logger,
		pathCh:  make(chan *PathInfo, 16),
		vniMap:  make(map[int][]int),
		msgBook: make(map[string]int64),
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
		}).Errorf("Parse rt failed %s", err.Error())
		return
	}
	// Get L2VNI from rd, find interface with rd, if exist add it to vnimap
	_, rd, err := utils.ParseVni(p.RD)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"Topic": "BUM",
		}).Errorf("Parse rd failed %s", err.Error())
		return
	}

	vtep, err := utils.GetVtepByVNI(rd)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"Topic": "BUM",
		}).Errorf("Get vtep by vni [%d] failed", rd)
		return
	}
	if len(vtep) == 0 {
		s.logger.WithFields(logrus.Fields{
			"Topic": "BUM",
		}).Errorf("No vxlan found with vni [%d], invalid bgp msg %s", rt, p.Detail())
		return
	}
	s.vniMap[rt] = append(s.vniMap[rt], rd)
}

func (s *AgentServer) handleDVR(vxlanInfos []*vxlanInfo) bool {
	for _, info := range vxlanInfos {
		s.logger.WithFields(logrus.Fields{
			"Topic": "DVR Vxlan Info",
		}).Info(info.Detail())

		// Parse cidr by dvr ip address
		cidr, err := utils.ParseNetworkInfo(info.ip, info.netLen, info.subnetLen)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"Topic": "DVR Vxlan Info",
			}).Errorf("Parse network info failed %s", err.Error())
			return false
		}
		fmt.Print(cidr.String())

		// Setup or teardown

		// Sync fdb, neigh, route
	}
	return true
}

func (s *AgentServer) handleVM(vxlanInfos []*vxlanInfo) bool {
	fmt.Println("Not Implement")
	return true
}

func (s *AgentServer) generateVxlanInfo(rt, rd int, p *PathInfo) []*vxlanInfo {
	var vxlanInfos []*vxlanInfo
	// Get all local vni list by rt, and get vtep by vni
	for _, localVni := range s.vniMap[rt] {
		vtep, err := utils.GetVtepByVNI(localVni)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"Topic": "VxlanInfo",
			}).Errorf("Get vtep by vni failed %s", err.Error())
			continue
		}
		dvrIface := utils.GetDvrIfaceByVtep(vtep)
		netnsId, err := utils.GetNetnsByIface(dvrIface)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"Topic": "VxlanInfo",
			}).Errorf("Get netns id of interface [%s] failed %s", dvrIface, err.Error())
			continue
		}

		netLen, subnetLen, err := utils.GetNetInfoFromComms(p.Communities)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"Topic": "VxlanInfo",
			}).Error(err.Error())
		}

		info := new(vxlanInfo)
		info.vni = rd
		info.ip = p.IP
		info.mac = p.Mac
		info.nexthop = p.Nexthop
		info.vtep = vtep
		info.netnsId = netnsId
		info.withdrawal = p.Withdrawal
		info.dvrIface = dvrIface
		info.netLen = netLen
		info.subnetLen = subnetLen
		vxlanInfos = append(vxlanInfos, info)
	}
	return vxlanInfos
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
	// Wait s.vniMap not empty before consume MACADV
	if len(s.vniMap) == 0 {
		// Put pathInfo back, nor it will block comsume BUM
		s.pathCh <- p
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

		// Check age before consume msg, only consume latest msg
		bookID := p.Identification()
		if bookInfo, ok := s.msgBook[bookID]; ok {
			if bookInfo >= p.Age {
				s.logger.WithFields(logrus.Fields{
					"Topic":  "MACADV",
					"Staled": true,
				}).Info(p.Detail())
				return
			}
		}

		// Generate vxlanInfo, sync fdb,route and neigh to all vtep under the same vpc
		vxlanInfos := s.generateVxlanInfo(rt, rd, p)

		switch vniType {
		case utils.MSG_TYPE_DVR:
			if s.handleDVR(vxlanInfos) {
				s.msgBook[bookID] = p.Age
			}
		case utils.MSG_TYPE_VM:
			if s.handleVM(vxlanInfos) {
				s.msgBook[bookID] = p.Age
			}
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

		// Handle MSG
		if p.IsBUM {
			if !p.IsLocal {
				s.logger.WithFields(logrus.Fields{
					"Topic": "Skip BUM",
				}).Debug(p.Detail())
				continue
			}
			s.handleBUM(p)
		} else {
			if p.IsLocal {
				s.logger.WithFields(logrus.Fields{
					"Topic": "Skip MACADV",
				}).Debug(p.Detail())
				continue
			}
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
