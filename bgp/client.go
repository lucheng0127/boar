package bgp

import (
	"context"
	"fmt"
	"io"
	"time"

	api "github.com/osrg/gobgp/v3/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getNeighbors(ctx context.Context, client api.GobgpApiClient, address string, enableAdv bool) ([]*api.Peer, error) {
	stream, err := client.ListPeer(ctx, &api.ListPeerRequest{
		Address:          address,
		EnableAdvertised: enableAdv,
	})
	if err != nil {
		return nil, err
	}

	l := make([]*api.Peer, 0, 1024)
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		l = append(l, r.Peer)
	}
	if address != "" && len(l) == 0 {
		return l, fmt.Errorf("not found neighbor %s", address)
	}
	return l, err
}

func newClient(ctx context.Context, host string, port uint32) (api.GobgpApiClient, context.CancelFunc, error) {
	grpcOpts := []grpc.DialOption{grpc.WithBlock()}
	grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	cc, cancel := context.WithTimeout(ctx, time.Second)
	target := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.DialContext(cc, target, grpcOpts...)
	if err != nil {
		return nil, cancel, err
	}
	return api.NewGobgpApiClient(conn), cancel, nil
}
