package main

import (
	"os"
	"os/signal"
	"syscall"

	hlm_settings "github.com/number571/go-peer/cmd/hidden_lake/messenger/internal/settings"
	"github.com/number571/go-peer/internal/pprof"
)

func main() {
	pprof.RunPprofService(hlm_settings.CServiceName)

	app, err := initApp()
	if err != nil {
		panic(err)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	if err := app.Run(); err != nil {
		panic(err)
	}
	defer func() {
		if err := app.Stop(); err != nil {
			panic(err)
		}
	}()

	<-shutdown
}
