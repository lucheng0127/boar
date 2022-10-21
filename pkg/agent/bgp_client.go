package agent

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	api "github.com/osrg/gobgp/v3/api"
	"github.com/osrg/gobgp/v3/pkg/apiutil"
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
}, pathCh chan *apiutil.Path) {
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
				pathCh <- &apiutil.Path{
					Nlri:       nlri,
					Age:        p.Age.AsTime().Unix(),
					Best:       p.Best,
					Attrs:      attrs,
					Stale:      p.Stale,
					Withdrawal: p.IsWithdraw,
					SourceID:   net.ParseIP(p.SourceId),
					NeighborIP: net.ParseIP(p.NeighborIp),
				}
			}
		}
	}
}
