package main

import (
	"flag"
	"fmt"
	"github.com/sestack/smq/broker"
	"github.com/sestack/smq/global"
	"github.com/sestack/smq/initialize"
	"os"
	"os/signal"
	"runtime"
)

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name x-token
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var (
		err error
	)

	cfg := flag.String("c", "./config.yaml", "configuration file")
	version := flag.Bool("v", false, "show version")

	flag.Parse()

	if *version {
		fmt.Printf("version %s\n", broker.Version)
		os.Exit(0)
	}

	global.VP, err = initialize.InitConfig(*cfg)
	if err != nil {
		fmt.Printf("failed to parse configuration file error:%v\n", err)
		os.Exit(1)
	}

	global.LOGGER, err = initialize.InitZap()
	if err != nil {
		fmt.Printf("failed to initialize log component error:%v\n", err)
		os.Exit(1)
	}

	broker.BROKER, err = broker.NewBroker(global.CONFIG)
	if err != nil {
		fmt.Printf("failed to initialize Broker error：%v\n", err)
		os.Exit(1)
	}

	broker.BROKER.Start()

	s := waitForSignal()
	fmt.Println("signal received, broker closed.", s)
}

func waitForSignal() os.Signal {
	signalChan := make(chan os.Signal, 1)
	defer close(signalChan)
	signal.Notify(signalChan, os.Kill, os.Interrupt)
	s := <-signalChan
	signal.Stop(signalChan)
	return s
}
