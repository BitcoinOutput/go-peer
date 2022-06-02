package network

import (
	"sync"

	"github.com/number571/go-peer/encoding"
	"github.com/number571/go-peer/local"
	"github.com/number571/go-peer/settings"
)

var (
	_ iOnline = &sOnline{}
)

type sOnline struct {
	fMutex   sync.Mutex
	fNode    INode
	fEnabled bool
}

func (onl *sOnline) Switch(state bool) {
	onl.fMutex.Lock()
	defer onl.fMutex.Unlock()

	if onl.fEnabled == state {
		return
	}
	onl.fEnabled = state

	switch state {
	case true:
		onl.start()
	case false:
		onl.stop()
	}
}

func (onl *sOnline) Status() bool {
	onl.fMutex.Lock()
	defer onl.fMutex.Unlock()

	return onl.fEnabled
}

func (onl *sOnline) start() {
	go func(node INode) {
		sett := node.Client().Settings()
		patt := encoding.Uint64ToBytes(sett.Get(settings.MaskPing))

		node.Handle(patt, func(node INode, msg local.IMessage) []byte {
			return patt
		})
	}(onl.fNode)
}

func (onl *sOnline) stop() {
	go func(node INode) {
		sett := node.Client().Settings()
		patt := encoding.Uint64ToBytes(sett.Get(settings.MaskPing))

		node.Handle(patt, nil)
	}(onl.fNode)
}