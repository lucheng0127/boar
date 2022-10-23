package agent

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/lucheng0127/boar/pkg/utils"
	api "github.com/osrg/gobgp/v3/api"
	"github.com/osrg/gobgp/v3/pkg/apiutil"
	"github.com/osrg/gobgp/v3/pkg/packet/bgp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient(ctx context.Context, host string) (api.GobgpApiClient, context.CancelFunc, error) {
	grpcOpts := []grpc.DialOption{grpc.WithBlock()}
	grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	cc, cancel := context.WithTimeout(ctx, time.Second)
	conn, err := grpc.DialContext(cc, host, grpcOpts...)
	if err != nil {
		return nil, cancel, err
	}
	return api.NewGobgpApiClient(conn), cancel, nil
}

func Monitor(recver interface {
	Recv() (*api.WatchEventResponse, error)
}, s *AgentServer) {
	for {
		r, err := recver.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(fmt.Sprintf("Monitor rib Failed %s", err.Error()))
		}
		if t := r.GetTable(); t != nil {
			for _, p := range t.Paths {
				if p.Family.Afi != api.Family_AFI_L2VPN {
					// Only handle evpn msg
					continue
				}
				nlri, _ := apiutil.GetNativeNlri(p)
				attrs, _ := apiutil.GetNativePathAttributes(p)
				pathInfo := new(PathInfo)

				evpnNlri := nlri.(*bgp.EVPNNLRI).RouteTypeData
				switch route := evpnNlri.(type) {
				case *bgp.EVPNMacIPAdvertisementRoute:
					pathInfo.RD = route.RD.String()
					pathInfo.IP = route.IPAddress
					pathInfo.Mac = route.MacAddress
				case *bgp.EVPNMulticastEthernetTagRoute:
					pathInfo.IsBUM = true
					pathInfo.RD = route.RD.String()
					pathInfo.IP = route.IPAddress
				}
				pathInfo.Age = p.Age.AsTime().Unix()
				pathInfo.Withdrawal = p.IsWithdraw
				pathInfo.Neighbor = net.ParseIP(p.NeighborIp)
				for _, attr := range attrs {
					switch a := attr.(type) {
					case *bgp.PathAttributeNextHop:
						pathInfo.Nexthop = a.Value
					case *bgp.PathAttributeMpReachNLRI:
						pathInfo.Nexthop = a.Nexthop
					case *bgp.PathAttributeAsPath:
						pathInfo.AsPath = a.String()
					case *bgp.PathAttributeCommunities:
						pathInfo.Communities = append(pathInfo.Communities, a.Value...)
					case *bgp.PathAttributeExtendedCommunities:
						extComms := a.Value
						for _, extComm := range extComms {
							extCommType, extCommSubType := extComm.GetTypes()
							if extCommType == bgp.EC_TYPE_TRANSITIVE_TWO_OCTET_AS_SPECIFIC && extCommSubType == bgp.EC_SUBTYPE_ROUTE_TARGET {
								pathInfo.RT = extComm.String()
							}
						}
					}
				}
				if isLocal, err := utils.IsLocalIP(pathInfo.Nexthop.String()); err != nil {
					s.logger.WithFields(logrus.Fields{
						"Topic": "Config",
					}).Errorf("Failed to get local IP address")
					panic(err)
				} else {
					pathInfo.IsLocal = isLocal
				}
				s.pathCh <- pathInfo
			}
		}
	}
}
