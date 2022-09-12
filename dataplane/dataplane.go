package dataplane

import (
	"sync"

	"github.com/lucheng0127/boar/bgp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Dataplane struct {
	routerID string
	port     uint32
	rrID     string
	vni      uint32
	asn      uint16
	cidr     string
	mtu      int
}

func NewDataplane() *Dataplane {
	return &Dataplane{
		routerID: viper.GetString("bgp.router_id"),
		port:     viper.GetUint32("bgp.listen_port"),
		rrID:     viper.GetString("bgp.rr_id"),
		asn:      uint16(viper.GetInt("bgp.as_num")),
		vni:      viper.GetUint32("bgp.vni"),
		cidr:     viper.GetString("bgp.cidr"),
		mtu:      viper.GetInt("bgp.mtu"),
	}
}

func (d *Dataplane) Serve() {
	log.Infof("Launch dataplane [%+v]", *d)

	checkInterfaces(d.vni, d.mtu, d.cidr, true)
	bgpServer := bgp.NewBGPServer(d.routerID, d.port, uint32(d.asn))
	var wg sync.WaitGroup
	wg.Add(1)
	go bgpServer.Serve(&wg)
	wg.Wait()
}
