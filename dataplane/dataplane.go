package dataplane

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Dataplane struct {
	vni    uint32
	as     uint16
	rrHost string
}

func NewDataplane() *Dataplane {
	vni := viper.GetUint32("bgp.vni")
	as := uint16(viper.GetInt("bgp.as_num"))
	rrHost := viper.GetString(("bgp.host"))
	return &Dataplane{
		vni:    vni,
		as:     as,
		rrHost: rrHost,
	}
}

func (d *Dataplane) Serve() {
	log.Infof("Launch dataplane [%+v]", *d)

}
