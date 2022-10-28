package utils

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/osrg/gobgp/v3/pkg/packet/bgp"
)

const (
	MSG_TYPE_NONE int = iota
	MSG_TYPE_VM
	MSG_TYPE_DVR
	CLS_FORMAT string = "2006-01-02 15:04:05"
)

var gLocalIP string

func getLocalIP() (string, error) {
	if len(gLocalIP) > 0 {
		return gLocalIP, nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	f, err := os.OpenFile("/etc/hosts", os.O_RDONLY, os.ModePerm)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scaner := bufio.NewScanner(f)
	for scaner.Scan() {
		if strings.Contains(scaner.Text(), hostname) {
			gLocalIP = strings.Split(scaner.Text(), " ")[0]
			break
		}
	}
	if err = scaner.Err(); err != nil {
		return "", err
	}
	return gLocalIP, nil
}

func IsLocalIP(ip string) (bool, error) {
	localIP, err := getLocalIP()
	if err != nil {
		return false, err
	}
	return ip == localIP, nil
}

func GetNextHopFromPathAttributes(attrs []bgp.PathAttributeInterface) net.IP {
	for _, attr := range attrs {
		switch a := attr.(type) {
		case *bgp.PathAttributeNextHop:
			return a.Value
		case *bgp.PathAttributeMpReachNLRI:
			return a.Nexthop
		}
	}
	return nil
}

func ParseVni(vni string) (int, int, error) {
	vniFragments := strings.Split(vni, ":")
	if len(vniFragments) != 2 {
		return 0, 0, fmt.Errorf("wrong vni format [%s]", vni)
	}
	vniType, err := strconv.Atoi(vniFragments[0])
	if err != nil {
		return 0, 0, err
	}
	vniValue, err := strconv.Atoi(vniFragments[1])
	if err != nil {
		return 0, 0, err
	}
	return vniType, vniValue, nil
}

func GetDvrIfaceByVtep(vtep string) string {
	return "kr_" + vtep[1:]
}

func GetNetInfoFromComms(comms []uint32) (int, int, error) {
	if len(comms) != 3 {
		return 0, 0, errors.New("wrong communites")
	}
	netLen := int(comms[2]) - int(comms[0])
	subnetLen := int(comms[1]) - int(comms[0])
	return netLen, subnetLen, nil
}

func ParseNetworkInfo(ip net.IP, netLen, subnetLen int) (*net.IPNet, error) {
	/*
		Input                |Output
		172.17.254.65 18 16  |172.17.64.0/18
		192.168.254.5 22 16  |192.168.4.0/22
		172.17.254.124 24 16 |172.17.123.0/24
		172.17.255.254 24 16 |0.0.0.0/0
		A.B.C.D X 16         |A.B.D-1.0/X
		A.B.255.254 X 16     |0.0.0.0/0
	*/
	if netLen > subnetLen || subnetLen > 32 {
		return nil, fmt.Errorf("invalid network length [%d] subnet network length [%d]", netLen, subnetLen)
	}

	ip = ip.To4()
	if ip == nil {
		return nil, errors.New("only ipv4 supported")
	}

	if ip[3] == 254 {
		return &net.IPNet{
			IP:   net.ParseIP("0.0.0.0"),
			Mask: net.CIDRMask(0, 32),
		}, nil
	}

	var cidr uint32
	cidr0 := ip[0]
	cidr1 := ip[1]
	cidr2 := ip[3] - 1
	cidr += uint32(cidr0) << 24
	cidr += uint32(cidr1) << 16
	cidr += uint32(cidr2) << 8

	cidrBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(cidrBytes, cidr)
	return &net.IPNet{
		IP:   cidrBytes,
		Mask: net.CIDRMask(subnetLen, 32),
	}, nil
}
