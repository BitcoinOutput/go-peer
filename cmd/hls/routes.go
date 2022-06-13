package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	hlsnet "github.com/number571/go-peer/cmd/hls/network"
	hls_settings "github.com/number571/go-peer/cmd/hls/settings"
	"github.com/number571/go-peer/crypto/hashing"
	"github.com/number571/go-peer/local/message"
	"github.com/number571/go-peer/local/routing"
	"github.com/number571/go-peer/network"
)

func routeHLS(node network.INode, msg message.IMessage) []byte {
	// load request from message's body
	requestBytes := msg.Body().Data()
	request := hlsnet.LoadRequest(requestBytes)
	if request == nil {
		return nil
	}

	// check already received data by hash
	hash := hashing.NewSHA256Hasher(requestBytes).Bytes()
	if gDB.Exist(hash) {
		return nil
	}
	gDB.Push(hash)

	// get service's address by name
	info, ok := gConfig.GetService(request.Host())
	if !ok {
		return nil
	}

	// redirect bytes of request to another nodes
	if info.IsRedirect() {
		for _, recv := range gConfig.F2F().PubKeys() {
			go node.Request(
				routing.NewRoute(recv),
				message.NewMessage([]byte(hls_settings.CTitlePattern), requestBytes),
			)
		}
	}

	// generate new request to serivce
	req, err := http.NewRequest(
		request.Method(),
		fmt.Sprintf("http://%s%s", info.Address(), request.Path()),
		bytes.NewReader(request.Body()),
	)
	if err != nil {
		return nil
	}
	for key, val := range request.Head() {
		req.Header.Add(key, val)
	}

	// send request to service
	// and receive response from service
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	// send result to client
	return data
}
