package network

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/number571/go-peer/local/payload"
	"github.com/number571/go-peer/settings"
	"github.com/number571/go-peer/storage"
)

var (
	_ INode = &sNode{}
)

// Basic structure for network use.
type sNode struct {
	fMutex        sync.Mutex
	fListener     net.Listener
	fSettings     settings.ISettings
	fHashMapping  storage.IKeyValueStorage
	fConnections  map[string]IConn
	fHandleRoutes map[uint64]IHandlerF
}

// Create client by private key as identification.
func NewNode(sett settings.ISettings) INode {
	sizeMapp := sett.Get(settings.CSizeMapp)
	return &sNode{
		fSettings:     sett,
		fHashMapping:  storage.NewMemoryStorage(sizeMapp),
		fConnections:  make(map[string]IConn),
		fHandleRoutes: make(map[uint64]IHandlerF),
	}
}

func (node *sNode) Broadcast(pl payload.IPayload) error {
	// set this message to mapping
	msg := NewMessage(pl)
	node.inMappingWithSet(msg.Hash())

	var err error
	for _, conn := range node.Connections() {
		e := conn.Write(msg)
		if e != nil {
			err = e
		}
	}

	return err
}

// Turn on listener by address.
// Client handle function need be not null.
func (node *sNode) Listen(address string) error {
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listen.Close()

	node.fListener = listen
	for {
		conn, err := listen.Accept()
		if err != nil {
			break
		}

		node.fMutex.Lock()
		isConnLimit := node.hasMaxConnSize()
		node.fMutex.Unlock()

		if isConnLimit {
			conn.Close()
			continue
		}

		node.fMutex.Lock()
		iconn := LoadConn(node.fSettings, conn)
		node.fConnections[iconn.Socket().RemoteAddr().String()] = iconn
		node.fMutex.Unlock()

		go node.handleConn(iconn)
	}

	return nil
}

func (node *sNode) Close() error {
	var err error

	node.fMutex.Lock()
	for id, conn := range node.fConnections {
		e := conn.Close()
		if e != nil {
			err = e
		}
		delete(node.fConnections, id)
	}
	if node.fListener != nil {
		e := node.fListener.Close()
		if e != nil {
			err = e
		}
	}
	node.fMutex.Unlock()

	return err
}

// Add function to mapping for route use.
func (node *sNode) Handle(head uint64, handle IHandlerF) INode {
	node.fMutex.Lock()
	defer node.fMutex.Unlock()

	node.fHandleRoutes[head] = handle
	return node
}

func (node *sNode) handleConn(conn IConn) {
	defer node.Disconnect(conn)

	var (
		retrySize = node.fSettings.Get(settings.CSizeRtry)
		msgNum    = node.fSettings.Get(settings.CSizeBmsg)

		retryCounter = uint64(0)
		msgCounter   = int64(0)
	)

	for {
		if atomic.LoadUint64(&retryCounter) > retrySize {
			break
		}

		if uint64(atomic.LoadInt64(&msgCounter)) > msgNum {
			time.Sleep(200 * time.Millisecond)
			continue
		}

		msg := conn.Read()
		go func(msg IMessage) {
			atomic.AddInt64(&msgCounter, 1)
			defer atomic.AddInt64(&msgCounter, -1)

			ok := node.handleMessage(conn, msg)
			if ok {
				atomic.StoreUint64(&retryCounter, 0)
				return
			}

			atomic.AddUint64(&retryCounter, 1)
		}(msg)
	}
}

// Get list of connection addresses.
func (node *sNode) Connections() []IConn {
	node.fMutex.Lock()
	defer node.fMutex.Unlock()

	var list []IConn
	for _, conn := range node.fConnections {
		list = append(list, conn)
	}

	return list
}

// Connect to node by address.
// Client handle function need be not null.
func (node *sNode) Connect(address string) IConn {
	node.fMutex.Lock()
	defer node.fMutex.Unlock()

	if node.hasMaxConnSize() {
		return nil
	}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil
	}

	iconn := LoadConn(node.fSettings, conn)
	node.fConnections[iconn.Socket().RemoteAddr().String()] = iconn

	go node.handleConn(iconn)

	return iconn
}

func (node *sNode) Disconnect(conn IConn) error {
	node.fMutex.Lock()
	defer node.fMutex.Unlock()

	delete(node.fConnections, conn.Socket().RemoteAddr().String())
	return conn.Close()
}

func (node *sNode) handleMessage(conn IConn, msg IMessage) bool {
	// null message from connection is error
	if msg == nil {
		return false
	}

	// check message in mapping by hash
	if node.inMappingWithSet(msg.Hash()) {
		return true
	}

	// get function by head
	f, ok := node.getFunction(msg.Payload().Head())
	if !ok {
		return false
	}

	f(node, conn, msg.Payload())
	return true
}

func (node *sNode) hasMaxConnSize() bool {
	maxConns := node.fSettings.Get(settings.CSizeConn)
	return uint64(len(node.fConnections)) > maxConns
}

func (node *sNode) inMappingWithSet(hash []byte) bool {
	node.fMutex.Lock()
	defer node.fMutex.Unlock()

	// skey already exists
	_, err := node.fHashMapping.Get(hash)
	if err == nil {
		return true
	}

	// push skey to mapping
	node.fHashMapping.Set(hash, []byte{1})
	return false
}

func (node *sNode) getFunction(head uint64) (IHandlerF, bool) {
	node.fMutex.Lock()
	defer node.fMutex.Unlock()

	f, ok := node.fHandleRoutes[head]
	return f, ok
}
