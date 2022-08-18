package api

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type APIServer struct {
	port uint16
}

func NewApiServer() *APIServer {
	port := uint16(viper.GetInt("api.port"))
	return &APIServer{
		port: port,
	}
}

func (a *APIServer) Serve() {
	log.Infof("Launch dataplane [%+v]", *a)
}
