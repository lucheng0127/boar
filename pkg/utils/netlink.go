package utils

import (
	"github.com/vishvananda/netlink"
)

func GetVtepByVNI(vni int) (string, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return "", err
	}

	var vtep string
	for _, link := range links {
		switch l := link.(type) {
		case *netlink.Vxlan:
			if l.VxlanId == vni {
				attrs := link.Attrs()
				vtep = attrs.Name
			}
		}
	}
	return vtep, nil
}
