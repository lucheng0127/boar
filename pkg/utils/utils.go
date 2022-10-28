package utils

import (
	"bufio"
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
	netLen := int(comms[1]) - int(comms[0])
	subnetLen := int(comms[2]) - int(comms[0])
	return netLen, subnetLen, nil
}

func ParseNetworkInfo(ip net.IP, netLen, subnetLen int) (net.IPNet, net.IP, error) {
	// TODO(shawnlu): Implement it
	return net.IPNet{}, net.IPv4(8, 8, 8, 8), nil
}
