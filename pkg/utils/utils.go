package utils

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/osrg/gobgp/v3/pkg/apiutil"
	"github.com/osrg/gobgp/v3/pkg/packet/bgp"
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

func PathDetail(path *apiutil.Path) string {
	t := time.Unix(path.Age, 0)
	tStr := t.UTC().Format(time.RFC3339)
	return fmt.Sprintf(
		"age %s neigh [%s] source [%s] withdraw [%t] nlri [%s]",
		tStr,
		path.NeighborIP.String(),
		path.SourceID.String(),
		path.Withdrawal,
		path.Nlri.String(),
	)
}
