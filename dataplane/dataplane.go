package dataplane

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Dataplane struct {
	vni    uint32
	as     uint16
	rrHost string
	cidr   string
	mtu    int
}

func NewDataplane() *Dataplane {
	return &Dataplane{
		vni:    viper.GetUint32("bgp.vni"),
		as:     uint16(viper.GetInt("bgp.as_num")),
		rrHost: viper.GetString("bgp.host"),
		cidr:   viper.GetString("bgp.cidr"),
		mtu:    viper.GetInt("bgp.mtu"),
	}
}

func (d *Dataplane) Serve() {
	log.Infof("Launch dataplane [%+v]", *d)

	checkInterfaces(d.vni, d.mtu, d.cidr, true)
}
