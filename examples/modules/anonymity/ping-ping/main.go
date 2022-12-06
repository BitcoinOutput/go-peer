package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/number571/go-peer/modules/client"
	"github.com/number571/go-peer/modules/crypto/asymmetric"
	"github.com/number571/go-peer/modules/friends"
	"github.com/number571/go-peer/modules/network"
	"github.com/number571/go-peer/modules/network/anonymity"
	payload_adapter "github.com/number571/go-peer/modules/network/anonymity/adapters/payload"
	"github.com/number571/go-peer/modules/network/conn"
	"github.com/number571/go-peer/modules/payload"
	"github.com/number571/go-peer/modules/queue"
	"github.com/number571/go-peer/modules/storage/database"
)

const (
	serviceHeader  = 0xDEADBEAF
	serviceAddress = ":8080"
)

const (
	dbPath1 = "database1.db"
	dbPath2 = "database2.db"
)

func deleteDBs() {
	os.RemoveAll(dbPath1)
	os.RemoveAll(dbPath2)
}

func main() {
	deleteDBs()
	defer deleteDBs()

	var (
		service1 = newNode(dbPath1)
		service2 = newNode(dbPath2)
	)

	service1.Handle(serviceHeader, handler("#1"))
	service2.Handle(serviceHeader, handler("#2"))

	service1.F2F().Append(service2.Queue().Client().PubKey())
	service2.F2F().Append(service1.Queue().Client().PubKey())

	if err := service1.Run(); err != nil {
		panic(err)
	}

	if err := service2.Run(); err != nil {
		panic(err)
	}

	go service1.Network().Listen(serviceAddress)
	time.Sleep(time.Second)

	if _, err := service2.Network().Connect(serviceAddress); err != nil {
		panic(err)
	}

	msg, err := service2.Queue().Client().Encrypt(
		service1.Queue().Client().PubKey(),
		payload_adapter.NewPayload(
			serviceHeader,
			[]byte("0"),
		),
	)
	if err != nil {
		panic(err)
	}

	if err := service2.Queue().Enqueue(msg); err != nil {
		panic(err)
	}

	select {}
}

func handler(serviceName string) anonymity.IHandlerF {
	return func(node anonymity.INode, pubKey asymmetric.IPubKey, pld payload.IPayload) []byte {
		num, err := strconv.Atoi(string(pld.Body()))
		if err != nil {
			panic(err)
		}

		val := "ping"
		if num%2 == 1 {
			val = "pong"
		}

		fmt.Printf("service '%s' got '%s#%d'\n", serviceName, val, num)

		msg, err := node.Queue().Client().Encrypt(
			pubKey,
			payload_adapter.NewPayload(
				serviceHeader,
				[]byte(fmt.Sprintf("%d", num+1)),
			),
		)
		if err != nil {
			panic(err)
		}

		if err := node.Queue().Enqueue(msg); err != nil {
			panic(err)
		}
		return nil
	}
}

func newNode(dbPath string) anonymity.INode {
	return anonymity.NewNode(
		anonymity.NewSettings(&anonymity.SSettings{}),
		database.NewLevelDB(
			database.NewSettings(&database.SSettings{}),
			dbPath,
		),
		network.NewNode(
			network.NewSettings(&network.SSettings{
				FConnSettings: conn.NewSettings(&conn.SSettings{}),
			}),
		),
		queue.NewQueue(
			queue.NewSettings(&queue.SSettings{}),
			client.NewClient(
				client.NewSettings(&client.SSettings{}),
				asymmetric.NewRSAPrivKey(1024),
			),
		),
		friends.NewF2F(),
	)
}