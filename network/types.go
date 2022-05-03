package network

import (
	"github.com/number571/go-peer/crypto"
	"github.com/number571/go-peer/local"
)

type iRouter func(INode) []crypto.IPubKey
type iHandler func(INode, local.IMessage) []byte

type INode interface {
	WithResponseRouter(iRouter) INode

	Request(local.IRoute, local.IMessage) ([]byte, error)
	Handle([]byte, iHandler) INode
	Listen(string) error
	Close()

	InConnections(string) bool
	Connections() []string
	Connect(string) error
	Disconnect(string)

	Client() local.IClient
	Checker() iChecker
	Pseudo() iPseudo
	Online() iOnline
	F2F() iF2F
}

type iOnline interface {
	iStatus
}

type iChecker interface {
	ListWithInfo() []iCheckerInfo

	iStatus
	iListPubKey
}

type iCheckerInfo interface {
	Online() bool
	PubKey() crypto.IPubKey
}

type iF2F interface {
	iStatus
	iListPubKey
}

type iListPubKey interface {
	InList(crypto.IPubKey) bool
	List() []crypto.IPubKey
	Append(crypto.IPubKey)
	Remove(crypto.IPubKey)
}

type iPseudo interface {
	iStatus
	Request(int) iPseudo
	Sleep() iPseudo
	PubKey() crypto.IPubKey
	PrivKey() crypto.IPrivKey
}

type iStatus interface {
	Switch(bool)
	Status() bool
}
