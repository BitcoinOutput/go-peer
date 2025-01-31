package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/number571/go-peer/cmd/hidden_lake/adapters/common"
	"github.com/number571/go-peer/internal/api"
)

func main() {
	if len(os.Args) != 3 {
		panic("./sender [port-incoming] [port-service]")
	}

	portIncoming, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}

	portService, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/traffic", trafficPage(portService))
	http.ListenAndServe(fmt.Sprintf(":%d", portIncoming), nil)
}

func trafficPage(portService int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			api.Response(w, 2, "failed: incorrect method")
			return
		}

		// get message from HLT
		msgBytes, err := io.ReadAll(r.Body)
		if err != nil {
			api.Response(w, 3, "failed: read body")
			return
		}

		ret, res := pushMessageToService(portService, msgBytes)
		api.Response(w, ret, res)
	}
}

func pushMessageToService(portService int, msgBytes []byte) (int, string) {
	// build request to service
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s:%d/push", common.HostService, portService),
		bytes.NewBuffer(msgBytes),
	)
	if err != nil {
		return 4, "failed: build request"
	}

	// send request to service
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 5, "failed: bad request"
	}
	defer resp.Body.Close()

	// read response from service
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return 6, "failed: read body from service"
	}

	// read body of response
	if len(res) == 0 || res[0] == '!' {
		return 7, "failed: incorrect response from service"
	}

	return 1, "success: push to service"
}
