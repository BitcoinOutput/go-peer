package app

import (
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/number571/go-peer/cmd/hls/config"
	"github.com/number571/go-peer/modules"
	"github.com/number571/go-peer/modules/closer"
	"github.com/number571/go-peer/modules/network/anonymity"
	"github.com/number571/go-peer/modules/network/conn_keeper"

	hls_settings "github.com/number571/go-peer/cmd/hls/settings"
)

var (
	_ IApp = &sApp{}
)

type sApp struct {
	fWrapper     config.IWrapper
	fNode        anonymity.INode
	fConnKeeper  conn_keeper.IConnKeeper
	fServiceHTTP *http.Server
}

func NewApp(
	cfg config.IConfig,
	node anonymity.INode,
) IApp {
	wrapper := config.NewWrapper(cfg)
	return &sApp{
		fWrapper:     wrapper,
		fNode:        node,
		fConnKeeper:  initConnKeeper(cfg, node),
		fServiceHTTP: initServiceHTTP(wrapper, node),
	}
}

func (app *sApp) Run() error {
	res := make(chan error)

	go func() {
		httpAddress := app.fWrapper.Config().Address().HTTP()
		if httpAddress == "" {
			return
		}
		err := app.fServiceHTTP.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			res <- err
			return
		}
	}()

	go func() {
		if err := app.fConnKeeper.Run(); err != nil {
			res <- err
			return
		}
	}()

	go func() {
		if err := app.fNode.Run(); err != nil {
			res <- err
			return
		}

		// if node in client mode
		// then run endless loop
		tcpAddress := app.fWrapper.Config().Address().TCP()
		if tcpAddress == "" {
			select {}
		}

		// run node in server mode
		err := app.fNode.Network().Listen(tcpAddress)
		if err != nil && !errors.Is(err, net.ErrClosed) {
			res <- err
			return
		}
	}()

	select {
	case err := <-res:
		app.Close()
		return err
	case <-time.After(time.Second * 3):
		return nil
	}
}

func (app *sApp) Close() error {
	return closer.CloseAll([]modules.ICloser{
		app.fNode,
		app.fConnKeeper,
		app.fServiceHTTP,
		app.fNode.Network(),
		app.fNode.KeyValueDB(),
	})
}

func initConnKeeper(cfg config.IConfig, node anonymity.INode) conn_keeper.IConnKeeper {
	return conn_keeper.NewConnKeeper(
		conn_keeper.NewSettings(&conn_keeper.SSettings{
			FConnections: func() []string { return cfg.Connections() },
			FDuration:    node.Settings().GetTimeWait(),
		}),
		node.Network(),
	)
}

func initServiceHTTP(wrapper config.IWrapper, node anonymity.INode) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc(hls_settings.CHandleRoot, handleIndexHTTP)
	mux.HandleFunc(hls_settings.CHandlePubKey, handlePubKeyHTTP(node))
	mux.HandleFunc(hls_settings.CHandleConnects, handleConnectsHTTP(wrapper, node))
	mux.HandleFunc(hls_settings.CHandleFriends, handleFriendsHTTP(wrapper, node))
	mux.HandleFunc(hls_settings.CHandleOnline, handleOnlineHTTP(node))
	mux.HandleFunc(hls_settings.CHandlePush, handlePushHTTP(node))

	return &http.Server{
		Addr:    wrapper.Config().Address().HTTP(),
		Handler: mux,
	}
}
