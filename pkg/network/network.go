package network

import (
	"fmt"
	"net"
	"sync"

	"github.com/number571/go-peer/pkg/crypto/hashing"
	"github.com/number571/go-peer/pkg/network/conn"
	"github.com/number571/go-peer/pkg/payload"
	"github.com/number571/go-peer/pkg/storage"
	"github.com/number571/go-peer/pkg/types"
)

var (
	_ INode = &sNode{}
)

type sNode struct {
	fMutex        sync.Mutex
	fListener     net.Listener
	fSettings     ISettings
	fHashMapping  storage.IKeyValueStorage
	fConnections  map[string]conn.IConn
	fHandleRoutes map[uint64]IHandlerF
}

// Creating a node object managed by connections with multiple nodes.
// Saves hashes of received messages to a buffer to prevent network cycling.
// Redirects messages to handle routers by keys.
func NewNode(pSett ISettings) INode {
	return &sNode{
		fSettings:     pSett,
		fHashMapping:  storage.NewMemoryStorage(pSett.GetCapacity()),
		fConnections:  make(map[string]conn.IConn),
		fHandleRoutes: make(map[uint64]IHandlerF),
	}
}

// Return settings interface.
func (p *sNode) GetSettings() ISettings {
	return p.fSettings
}

// Puts the hash of the message in the buffer and sends the message to all connections of the node.
func (p *sNode) BroadcastPayload(pPld payload.IPayload) error {
	hasher := hashing.NewSHA256Hasher(pPld.ToBytes())
	p.inMappingWithSet(hasher.ToBytes())

	var err error
	for _, conn := range p.GetConnections() {
		e := conn.WritePayload(pPld)
		if e != nil {
			err = e
		}
	}

	return err
}

// Opens a tcp connection to receive data from outside.
// Checks the number of valid connections.
// Redirects connections to the handle router.
func (p *sNode) Run() error {
	listener, err := net.Listen("tcp", p.GetSettings().GetAddress())
	if err != nil {
		return err
	}

	go func(pListener net.Listener) {
		defer pListener.Close()
		p.setListener(pListener)
		for {
			tconn, err := p.getListener().Accept()
			if err != nil {
				break
			}

			if p.hasMaxConnSize() {
				tconn.Close()
				continue
			}

			sett := p.GetSettings().GetConnSettings()
			conn := conn.LoadConn(sett, tconn)
			address := tconn.RemoteAddr().String()

			p.setConnection(address, conn)
			go p.handleConn(address, conn)
		}
	}(listener)

	return nil
}

// Closes the listener and all connections.
func (p *sNode) Stop() error {
	p.fMutex.Lock()
	defer p.fMutex.Unlock()

	toClose := make([]types.ICloser, 0, len(p.fConnections)+1)
	if p.fListener != nil {
		toClose = append(toClose, p.fListener)
	}

	for id, conn := range p.fConnections {
		toClose = append(toClose, conn)
		delete(p.fConnections, id)
	}

	return types.CloseAll(toClose)
}

// Saves the function to the map by key for subsequent redirection.
func (p *sNode) HandleFunc(pHead uint64, pHandle IHandlerF) INode {
	p.fMutex.Lock()
	defer p.fMutex.Unlock()

	p.fHandleRoutes[pHead] = pHandle
	return p
}

// Retrieves the entire list of connections with addresses.
func (p *sNode) GetConnections() map[string]conn.IConn {
	p.fMutex.Lock()
	defer p.fMutex.Unlock()

	var mapping = make(map[string]conn.IConn, len(p.fConnections))
	for addr, conn := range p.fConnections {
		mapping[addr] = conn
	}

	return mapping
}

// Connects to the node at the specified address and automatically starts reading all incoming messages.
// Checks the number of connections.
func (p *sNode) AddConnect(pAddress string) error {
	if p.hasMaxConnSize() {
		return fmt.Errorf("has max connections size")
	}

	sett := p.GetSettings().GetConnSettings()
	conn, err := conn.NewConn(sett, pAddress)
	if err != nil {
		return err
	}

	p.setConnection(pAddress, conn)
	go p.handleConn(pAddress, conn)

	return nil
}

// Disables the connection at the address and removes the connection from the connection list.
func (p *sNode) DelConnect(pAddress string) error {
	p.fMutex.Lock()
	defer p.fMutex.Unlock()

	conn, ok := p.fConnections[pAddress]
	if !ok {
		return nil
	}

	delete(p.fConnections, pAddress)
	return conn.Close()
}

// Processes the received data from the connection.
func (p *sNode) handleConn(pAddress string, pConn conn.IConn) {
	defer p.DelConnect(pAddress)
	for {
		ok := p.handleMessage(pConn, pConn.ReadPayload())
		if !ok {
			break
		}
	}
}

// Processes the message for correctness and redirects it to the handler function.
// Returns true if the message was successfully redirected to the handler function
// > or if the message already existed in the hash value store.
func (p *sNode) handleMessage(pConn conn.IConn, pPld payload.IPayload) bool {
	// null message from connection is error
	if pPld == nil {
		return false
	}

	// check message in mapping by hash
	hash := hashing.NewSHA256Hasher(pPld.ToBytes()).ToBytes()
	if p.inMappingWithSet(hash) {
		return true
	}

	// get function by head
	f, ok := p.getFunction(pPld.GetHead())
	if !ok || f == nil {
		return false
	}

	f(p, pConn, pPld.GetBody())
	return true
}

// Checks the current number of connections with the limit.
func (p *sNode) hasMaxConnSize() bool {
	p.fMutex.Lock()
	defer p.fMutex.Unlock()

	maxConns := p.GetSettings().GetMaxConnects()
	return uint64(len(p.fConnections)) > maxConns
}

// Checks the hash of the message for existence in the hash store.
// Returns true if the hash already existed, otherwise false.
func (p *sNode) inMappingWithSet(pHash []byte) bool {
	p.fMutex.Lock()
	defer p.fMutex.Unlock()

	// skey already exists
	_, err := p.fHashMapping.Get(pHash)
	if err == nil {
		return true
	}

	// push skey to mapping
	p.fHashMapping.Set(pHash, []byte{1})
	return false
}

// Saves the connection to the map.
func (p *sNode) setConnection(pAddress string, pConn conn.IConn) {
	p.fMutex.Lock()
	defer p.fMutex.Unlock()

	p.fConnections[pAddress] = pConn
}

// Gets the handler function by key.
func (p *sNode) getFunction(pHead uint64) (IHandlerF, bool) {
	p.fMutex.Lock()
	defer p.fMutex.Unlock()

	f, ok := p.fHandleRoutes[pHead]
	return f, ok
}

// Sets the listener.
func (p *sNode) setListener(pListener net.Listener) {
	p.fMutex.Lock()
	defer p.fMutex.Unlock()

	p.fListener = pListener
}

// Gets the listener.
func (p *sNode) getListener() net.Listener {
	p.fMutex.Lock()
	defer p.fMutex.Unlock()

	return p.fListener
}
