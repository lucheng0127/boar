package utils

import (
	"fmt"

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
	if len(vtep) == 0 {
		return "", fmt.Errorf("no vlxan vtep with vni [%d]", vni)
	}
	return vtep, nil
}

func GetNetnsByIface(iface string) (int, error) {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return 0, err
	}
	attrs := link.Attrs()
	return attrs.NetNsID, nil
}
