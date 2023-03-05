package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/oizgagin/ing/app"
	"github.com/oizgagin/ing/pkg/config"
)

var (
	configFile = flag.String("config", "/etc/ing/ing.toml", "path to config file")
)

func main() {
	flag.Parse()

	sigCh := make(chan os.Signal, 1)

	go func() {
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		<-sigCh
	}()

	var cfg app.Config

	err := config.ParseFile(*configFile, &cfg)
	if err != nil {
		log.Fatalf("could not parse config %v: %v", *configFile, err)
	}

	app, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("could not run app: %v", err)
	}

	<-sigCh

	if err := app.Close(); err != nil {
		log.Fatalf("could not close app: %v", err)
	}
}
