package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/lucheng0127/boar/api"
	"github.com/lucheng0127/boar/dataplane"
	"github.com/lucheng0127/boar/internal"
	"github.com/lucheng0127/boar/serve"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var cfgFile *string = flag.String("cfg", "/etc/boar/boar.yaml", "Config file")
	var debug *bool = flag.Bool("debug", false, "Enable debug")
	flag.Parse()

	internal.SetLogLevel(*debug)
	defer func() {
		if r := recover(); r != nil {
			log.Panicln(r)
			os.Exit(1)
		}
	}()

	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfgFile)
	log.Debugf("Read config from %s", *cfgFile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	apiServer := api.NewApiServer()
	go serve.Launch(apiServer)
	dataplane := dataplane.NewDataplane()
	go serve.Launch(dataplane)

	sig := <-sigCh
	log.Infof("Got %s signal, exit program ...", sig)
	os.Exit(-1)
}
