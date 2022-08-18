package dataplane

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"

	log "github.com/sirupsen/logrus"
)

const (
	VRF_ROUTE_TABLE = 10
)

func getVtepByVNI(vni uint32) (netlink.Link, error) {
	vtepName := getVtepName(vni)
	link, err := netlink.LinkByName(vtepName)
	if err != nil {
		return nil, err
	}

	if vtep, ok := link.(*netlink.Vxlan); ok {
		if vtep.VxlanId == int(vni) {
			return vtep, nil
		}
	}
	return nil, fmt.Errorf("vxlan link with vni [%d] not found", vni)
}

func setupInterfaces(vni uint32, mtu int, cidr string) error {
	log.Info("Createting local interfaces ...")

	handle, err := netlink.NewHandle()
	if err != nil {
		log.Errorf("New netlink handle\n%s", err.Error())
		return err
	}

	// Create vrf
	vrfAttrs := netlink.LinkAttrs{
		MTU:  mtu,
		Name: getVrfName(vni),
	}
	vrf := netlink.Vrf{
		LinkAttrs: vrfAttrs,
		Table:     uint32(VRF_ROUTE_TABLE),
	}
	if err := netlink.LinkAdd(&vrf); err != nil {
		log.Errorf("Create VRF\n%s", err.Error())
		return err
	}

	// Create vxlan vtep set vni
	vtepAttr := netlink.LinkAttrs{
		Name: getVtepName(vni),
	}
	vtep := netlink.Vxlan{
		LinkAttrs: vtepAttr,
		VxlanId:   int(vni),
		Port:      4789,
		Proxy:     true,
	}
	if err := netlink.LinkAdd(&vtep); err != nil {
		log.Errorf("Create Vxlan vtep\n%s", err.Error())
		return err
	}

	// Set vtep master vrf
	err = handle.LinkSetMasterByIndex(&vtep, vrf.Attrs().Index)
	if err != nil {
		log.Errorf("Link set master\n%s", err.Error())
		return err
	}

	// Set vtep and vrf up
	err = handle.LinkSetUp(&vrf)
	if err != nil {
		log.Errorf("Set vrf up\n%s", err.Error())
	}
	err = handle.LinkSetUp(&vtep)
	if err != nil {
		log.Errorf("Set vxlan vtep up\n%s", err.Error())
	}

	// Add ip route for vrf
	ipv4Addr, ipv4Net, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Errorf("Parse CIDR\n%s", err.Error())
		return err
	}
	vrfAddr := net.IPNet{
		IP:   ipv4Addr,
		Mask: ipv4Net.Mask,
	}
	err = netlink.AddrAdd(&vrf, &netlink.Addr{IPNet: &vrfAddr})
	if err != nil {
		log.Errorf("Set IP address for vrf\n%s", err.Error())
		return err
	}

	log.Info("Create local interfaces finished")
	return nil
}

func checkInterfaces(vni uint32, mtu int, cidr string, ensure bool) error {
	log.Info("Checking local interfaces ...")
	_, err := getVtepByVNI(vni)
	if err != nil {
		if !ensure {
			log.Errorf("Retrive vxlan vtep\n%s", err.Error())
			return err
		}
		err = setupInterfaces(vni, mtu, cidr)
		if err != nil {
			return err
		}
	}
	log.Info("Local interfaces check finished")
	return nil
}
