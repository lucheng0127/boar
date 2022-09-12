package bgp

import (
	"context"
	"sync"

	api "github.com/osrg/gobgp/v3/api"
	"github.com/osrg/gobgp/v3/pkg/server"
	log "github.com/sirupsen/logrus"
)

type BGPNeighbor struct {
	routerID string
	asn      uint32
}

type bgpServer struct {
	client    api.GobgpApiClient
	routerID  string
	port      uint32
	asn       uint32
	server    *server.BgpServer
	neighbors []BGPNeighbor
}

func NewBGPServer(routerID string, port, asn uint32) *bgpServer {
	bgpServer := new(bgpServer)
	bgpServer.routerID = routerID
	bgpServer.port = port
	bgpServer.asn = asn
	bgpServer.neighbors = make([]BGPNeighbor, 0)
	server := server.NewBgpServer()
	bgpServer.server = server
	return bgpServer
}

func (s *bgpServer) Serve(wg *sync.WaitGroup) {
	ctx := context.Background()
	client, cancel, err := newClient(ctx, s.routerID, s.port)
	if err != nil {
		panic(err)
	}
	defer cancel()
	s.client = client

	log.Info("Run bgp server on port [%d]", s.port)
}
