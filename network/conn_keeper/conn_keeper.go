package conn_keeper

import (
	"errors"
	"sync"
	"time"

	"github.com/number571/go-peer/network"
)

var (
	_ IConnKeeper = &sConnKeeper{}
)

type sConnKeeper struct {
	fMutex    sync.Mutex
	fEnable   bool
	fSignal   chan struct{}
	fNode     network.INode
	fSettings ISettings
}

func NewConnKeeper(sett ISettings, node network.INode) IConnKeeper {
	return &sConnKeeper{
		fSignal:   make(chan struct{}),
		fNode:     node,
		fSettings: sett,
	}
}

func (connKeeper *sConnKeeper) Settings() ISettings {
	return connKeeper.fSettings
}

func (connKeeper *sConnKeeper) Run() error {
	connKeeper.fMutex.Lock()
	defer connKeeper.fMutex.Unlock()

	if connKeeper.fEnable {
		return errors.New("conn keeper already enabled")
	}
	connKeeper.fEnable = true

	go func() {
		for {
			select {
			case <-connKeeper.fSignal:
				connKeeper.fEnable = false
				return
			default:
				connKeeper.tryConnectToAll()
				time.Sleep(connKeeper.Settings().GetDuration())
			}
		}
	}()

	return nil
}

func (connKeeper *sConnKeeper) Close() error {
	connKeeper.fMutex.Lock()
	defer connKeeper.fMutex.Unlock()

	if !connKeeper.fEnable {
		return errors.New("pull already disabled")
	}

	connKeeper.fSignal <- struct{}{}
	return nil
}

func (connKeeper *sConnKeeper) tryConnectToAll() {
	for _, address := range connKeeper.Settings().GetConnections() {
		for _, conn := range connKeeper.fNode.Connections() {
			connAddr := conn.Socket().RemoteAddr().String()
			if connAddr == address {
				continue
			}
		}
		conn := connKeeper.fNode.Connect(address)
		if conn == nil {
			continue
		}
	}
}