package netanon

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/number571/go-peer/crypto/asymmetric"
	"github.com/number571/go-peer/local/client"
	"github.com/number571/go-peer/local/payload"
	"github.com/number571/go-peer/local/selector"
	"github.com/number571/go-peer/settings"
)

var (
	tgAddrs = [2]string{":8081", ":9091"}
)

const (
	tcHead = 0xDEADBEAF
	tcIter = 10
	tcBody = "hello, world!"
)

func TestComplex(t *testing.T) {
	nodes := testNewNodes()
	defer testFreeNodes(nodes[:])

	for _, node := range nodes {
		node.Pseudo().Switch(true)
		node.Online().Switch(true)
		node.Checker().Switch(true)
	}

	wg := sync.WaitGroup{}
	wg.Add(tcIter)

	for i := 0; i < tcIter; i++ {
		go func(i int) {
			defer wg.Done()
			reqBody := fmt.Sprintf("%s (%d)", tcBody, i)

			// nodes[1] -> nodes[0] -> nodes[2]
			resp, err := nodes[0].Request(
				nodes[1].Client().PubKey(),
				payload.NewPayload(tcHead, []byte(reqBody)),
			)
			if err != nil {
				t.Errorf("%s (%d)", err.Error(), i)
				return
			}

			if string(resp) != fmt.Sprintf("%s (response)", reqBody) {
				t.Errorf("string(resp) != reqBody (%d)", i)
				return
			}
		}(i)
	}

	wg.Wait()
}

func TestPseudo(t *testing.T) {
	nodes := testNewNodes()
	defer testFreeNodes(nodes[:])

	for _, node := range nodes {
		node.Pseudo().Switch(true)
	}

	time.Sleep(1 * time.Second)
	reqBody := "hello, world!"

	resp, err := nodes[0].Request(
		nodes[1].Client().PubKey(),
		payload.NewPayload(tcHead, []byte(reqBody)),
	)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if string(resp) != fmt.Sprintf("%s (response)", reqBody) {
		t.Errorf("string(resp) != reqBody")
		return
	}
}

func TestOnlineChecker(t *testing.T) {
	nodes := testNewNodes()
	defer testFreeNodes(nodes[:])

	for _, node := range nodes {
		node.Online().Switch(true)
		node.Checker().Switch(true)
	}

	nodes[0].Checker().Append(nodes[1].Client().PubKey())
	time.Sleep(1 * time.Second)

	list := nodes[0].Checker().ListWithInfo()
	if !list[0].Online() {
		t.Errorf("node is offline")
		return
	}
}

func TestF2F(t *testing.T) {
	nodes := testNewNodes()
	defer testFreeNodes(nodes[:])

	testWithF2F(t, nodes, 1) // f2f with friends
}

func TestF2FWithoutFriends(t *testing.T) {
	nodes := testNewNodes()
	defer testFreeNodes(nodes[:])

	testWithF2F(t, nodes, 2) // f2f without friends
}

func testWithF2F(t *testing.T, nodes [5]INode, mode int) {
	nodes[0].F2F().Switch(true)
	nodes[1].F2F().Switch(true)

	switch mode {
	case 1:
		nodes[0].F2F().Append(nodes[1].Client().PubKey())
		nodes[1].F2F().Append(nodes[0].Client().PubKey())
	case 2:
		nodes[0].Client().Settings().Set(settings.CTimeWait, 1) // seconds
	default:
		// pass
	}

	reqBody := fmt.Sprintf("%s (%d)", tcBody, mode)

	// nodes[1] -> nodes[0] -> nodes[2]
	resp, err := nodes[0].Request(
		nodes[1].Client().PubKey(),
		payload.NewPayload(tcHead, []byte(reqBody)),
	)
	if err != nil {
		if mode == 2 {
			return
		}
		t.Errorf("%s (mode=%d)", err.Error(), mode)
		return
	}

	if string(resp) != fmt.Sprintf("%s (response)", reqBody) {
		t.Errorf("string(resp) != reqBody")
		return
	}
}

func testGetPubKeys(nodes []INode) []asymmetric.IPubKey {
	pubKeys := []asymmetric.IPubKey{}
	for _, node := range nodes {
		pubKeys = append(pubKeys, node.Client().PubKey())
	}
	return pubKeys
}

// nodes[0], nodex[1] = clients
// nodes[2], nodes[3], nodes[4] = routes
// (nodes[0]) -> nodes[2] -> nodes[3] -> nodes[4] -> (nodes[1])
func testNewNodes() [5]INode {
	nodes := [5]INode{}

	for i := 0; i < 5; i++ {
		nodes[i] = NewNode(testNewClient())
	}

	for _, node := range nodes {
		node.WithRouter(func() []asymmetric.IPubKey {
			return selector.NewSelector(
				testGetPubKeys(nodes[2:]),
			).Shuffle().Return(3)
		})

		node.Handle(
			settings.MustBeUint32(tcHead),
			func(node INode, sender asymmetric.IPubKey, pl payload.IPayload) []byte {
				// send response
				resp := fmt.Sprintf("%s (response)", string(pl.Body()))
				return []byte(resp)
			},
		)
	}

	go func() {
		err := nodes[2].Network().Listen(tgAddrs[0])
		if err != nil {
			panic(err)
		}
	}()
	go func() {
		err := nodes[4].Network().Listen(tgAddrs[1])
		if err != nil {
			panic(err)
		}
	}()

	time.Sleep(200 * time.Millisecond)

	// nodes to routes
	nodes[0].Network().Connect(tgAddrs[0])
	nodes[1].Network().Connect(tgAddrs[1])

	// routes to routes
	nodes[3].Network().Connect(tgAddrs[0])
	nodes[3].Network().Connect(tgAddrs[1])

	return nodes
}

func testFreeNodes(nodes []INode) {
	for _, node := range nodes {
		node.Close()
	}
}

func testNewClient() client.IClient {
	return client.NewClient(
		settings.NewSettings(),
		asymmetric.NewRSAPrivKey(1024),
	)
}