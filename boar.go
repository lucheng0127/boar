package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/lucheng0127/boar/internal/pkg/config"
	"github.com/lucheng0127/boar/pkg/server"

	flags "github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

const VERSION string = "0.1"

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var opts struct {
		ConfigFile    string `short:"f" long:"config-file" description:"config file"`
		ConfigType    string `short:"t" long:"config-type" description:"config file type (toml, yaml, json)" default:"yaml"`
		LogLevel      string `short:"l" long:"log-level" description:"log level"`
		DisableStdlog bool   `long:"disable-stdlog" description:"disable standard logging"`
		CPUs          int    `long:"cpus" description:"number of CPUs to be used"`
		Dry           bool   `short:"d" long:"dry-run" description:"check configuration"`
		Version       bool   `long:"version" description:"show version number"`
	}
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	if opts.Version {
		fmt.Println("Boar version ", VERSION)
	}

	if opts.CPUs == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	} else {
		if runtime.NumCPU() < opts.CPUs {
			logger.Errorf("Only %d CPUs are available, wanted %d", runtime.NumCPU(), opts.CPUs)
			os.Exit(1)
		}
		runtime.GOMAXPROCS(opts.CPUs)
	}

	switch opts.LogLevel {
	case "debug":
		logger.SetReportCaller(true)
		logger.SetLevel(logrus.DebugLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	if opts.DisableStdlog {
		logger.SetOutput(io.Discard)
	} else {
		logger.SetOutput(os.Stdout)
	}

	config, err := config.ReadConfigFile(opts.ConfigFile, opts.ConfigType)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"Topic": "Config",
			"Error": err,
		}).Fatalf("Can't read config file %s", opts.ConfigFile)
	}
	logger.WithFields(logrus.Fields{
		"Topic": "Config",
	}).Info("Finished reading config file")
	if opts.LogLevel == "debug" {
		fmt.Printf("%+v\n", config)
	}
	if opts.Dry {
		os.Exit(0)
	}

	logger.Info("Boar started")
	s := server.NewServer(config.Api.Port, config.Agent.Host, logger)
	go s.Serve()

	sig := <-sigCh
	logger.Infof("Got %s signal, exit program ...", sig)
}
