package netanon

import (
	"bytes"
	"sync"
	"time"

	"github.com/number571/go-peer/crypto/asymmetric"
	"github.com/number571/go-peer/encoding"
	"github.com/number571/go-peer/local/payload"
	"github.com/number571/go-peer/settings"
)

var (
	_ iChecker     = &sChecker{}
	_ iCheckerInfo = &sCheckerInfo{}
)

type sChecker struct {
	fMutex   sync.Mutex
	fNode    INode
	fEnabled bool
	fChannel chan struct{}
	fMapping map[string]iCheckerInfo
}

type sCheckerInfo struct {
	fMutex  sync.Mutex
	fOnline bool
	fPubKey asymmetric.IPubKey
}

func newChecker(node INode) iChecker {
	return &sChecker{
		fNode:    node,
		fChannel: make(chan struct{}),
		fMapping: make(map[string]iCheckerInfo),
	}
}

// Set state = bool.
func (checker *sChecker) Switch(state bool) {
	checker.fMutex.Lock()
	defer checker.fMutex.Unlock()

	if checker.fEnabled == state {
		return
	}
	checker.fEnabled = state

	switch state {
	case true:
		checker.start()
	case false:
		checker.stop()
	}
}

// Get current state of online mode.
func (checker *sChecker) Status() bool {
	checker.fMutex.Lock()
	defer checker.fMutex.Unlock()

	return checker.fEnabled
}

// Get a list of checks public keys and online status.
func (checker *sChecker) ListWithInfo() []iCheckerInfo {
	checker.fMutex.Lock()
	defer checker.fMutex.Unlock()

	var list []iCheckerInfo
	for _, chk := range checker.fMapping {
		list = append(list, chk)
	}

	return list
}

// Check the existence in the list by the public key.
func (checker *sChecker) InList(pub asymmetric.IPubKey) bool {
	checker.fMutex.Lock()
	defer checker.fMutex.Unlock()

	_, ok := checker.fMapping[pub.Address().String()]
	return ok
}

// Get a list of checks public keys.
func (checker *sChecker) List() []asymmetric.IPubKey {
	checker.fMutex.Lock()
	defer checker.fMutex.Unlock()

	var list []asymmetric.IPubKey
	for _, chk := range checker.fMapping {
		list = append(list, chk.PubKey())
	}

	return list
}

// Add public key to list of checks.
func (checker *sChecker) Append(pub asymmetric.IPubKey) {
	checker.fMutex.Lock()
	defer checker.fMutex.Unlock()

	checker.fMapping[pub.Address().String()] = &sCheckerInfo{
		fOnline: false,
		fPubKey: pub,
	}
}

// Delete public key from list of checks.
func (checker *sChecker) Remove(pub asymmetric.IPubKey) {
	checker.fMutex.Lock()
	defer checker.fMutex.Unlock()

	delete(checker.fMapping, pub.Address().String())
}

func (checkerInfo *sCheckerInfo) PubKey() asymmetric.IPubKey {
	return checkerInfo.fPubKey
}

func (checkerInfo *sCheckerInfo) Online() bool {
	checkerInfo.fMutex.Lock()
	defer checkerInfo.fMutex.Unlock()

	return checkerInfo.fOnline
}

func (checker *sChecker) start() {
	node := checker.fNode.(*sNode)
	sett := node.fClient.Settings()

	go func(checker *sChecker, node *sNode, sett settings.ISettings) {
		timeRslp := sett.Get(settings.CTimePing)
		for {
			broadcastOnline(checker, node, sett)
			select {
			case <-checker.fChannel:
				return
			case <-time.After(calcRandTimeInMS(1, timeRslp)):
				continue
			}
		}
	}(checker, node, sett)
}

func broadcastOnline(checker *sChecker, node *sNode, sett settings.ISettings) {
	list := checker.ListWithInfo()

	maskPing := sett.Get(settings.CMaskPing)
	maskPingBytes := encoding.Uint64ToBytes(maskPing)

	wg := sync.WaitGroup{}
	wg.Add(len(list))

	for _, recv := range list {
		go func(recv *sCheckerInfo) {
			defer wg.Done()
			resp, err := node.doRequest(
				recv.fPubKey,
				payload.NewPayload(maskPing, []byte{}),
				node.fRouterF,
				0, // retry number
				sett.Get(settings.CTimeWait),
			)
			if err != nil || !bytes.Equal(resp, maskPingBytes) {
				recv.fOnline = false
				return
			}
			recv.fOnline = true
		}(recv.(*sCheckerInfo))
	}

	wg.Wait()
}

func (checker *sChecker) stop() {
	checker.fChannel <- struct{}{}
}
