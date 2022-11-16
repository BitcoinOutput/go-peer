package hlc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	hls_settings "github.com/number571/go-peer/cmd/hls/settings"
	"github.com/number571/go-peer/modules/crypto/asymmetric"
	"github.com/number571/go-peer/modules/encoding"
)

var (
	_ IRequester = &sRequester{}
)

type sRequester struct {
	fHost string
}

func NewRequester(host string) IRequester {
	return &sRequester{
		fHost: host,
	}
}

func (requester *sRequester) Request(push *hls_settings.SPush) ([]byte, error) {
	res, err := doRequest(
		http.MethodPost,
		requester.fHost+hls_settings.CHandlePush,
		push,
	)
	if err != nil {
		return nil, err
	}

	return encoding.HexDecode(res), nil
}

func (requester *sRequester) Broadcast(push *hls_settings.SPush) error {
	_, err := doRequest(
		http.MethodPut,
		requester.fHost+hls_settings.CHandlePush,
		push,
	)
	return err
}

func (requester *sRequester) GetFriends() (map[string]asymmetric.IPubKey, error) {
	res, err := doRequest(
		http.MethodGet,
		requester.fHost+hls_settings.CHandleFriends,
		nil,
	)
	if err != nil {
		return nil, err
	}

	listFriends := deleteVoidStrings(strings.Split(res, ","))
	result := make(map[string]asymmetric.IPubKey, len(listFriends))
	for _, friend := range listFriends {
		splited := strings.Split(friend, ":")
		if len(splited) != 2 {
			return nil, fmt.Errorf("length of splited != 2")
		}
		aliasName := splited[0]
		pubKeyStr := splited[1]
		result[aliasName] = asymmetric.LoadRSAPubKey(pubKeyStr)
	}
	return result, nil
}

func (requester *sRequester) AddFriend(friend *hls_settings.SFriend) error {
	_, err := doRequest(
		http.MethodPost,
		requester.fHost+hls_settings.CHandleFriends,
		friend,
	)
	return err
}

func (requester *sRequester) DelFriend(friend *hls_settings.SFriend) error {
	_, err := doRequest(
		http.MethodDelete,
		requester.fHost+hls_settings.CHandleFriends,
		friend,
	)
	return err
}

func (requester *sRequester) GetOnlines() ([]string, error) {
	res, err := doRequest(
		http.MethodGet,
		requester.fHost+hls_settings.CHandleOnline,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return deleteVoidStrings(strings.Split(res, ",")), nil
}

func (requester *sRequester) DelOnline(connect *hls_settings.SConnect) error {
	_, err := doRequest(
		http.MethodDelete,
		requester.fHost+hls_settings.CHandleOnline,
		connect,
	)
	return err
}

func (requester *sRequester) GetConnections() ([]string, error) {
	res, err := doRequest(
		http.MethodGet,
		requester.fHost+hls_settings.CHandleConnects,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return deleteVoidStrings(strings.Split(res, ",")), nil
}

func (requester *sRequester) AddConnection(connect *hls_settings.SConnect) error {
	_, err := doRequest(
		http.MethodPost,
		requester.fHost+hls_settings.CHandleConnects,
		connect,
	)
	return err
}

func (requester *sRequester) DelConnection(connect *hls_settings.SConnect) error {
	_, err := doRequest(
		http.MethodDelete,
		requester.fHost+hls_settings.CHandleConnects,
		connect,
	)
	return err
}

func (requester *sRequester) PubKey() (asymmetric.IPubKey, error) {
	res, err := doRequest(
		http.MethodGet,
		requester.fHost+hls_settings.CHandlePubKey,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return asymmetric.LoadRSAPubKey(res), nil
}

func doRequest(method, url string, data interface{}) (string, error) {
	jsonValue, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(
		method,
		url,
		bytes.NewBuffer(jsonValue),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", hls_settings.CContentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	res, err := loadResponse(resp.Body)
	if err != nil {
		return "", err
	}

	return res.FResult, nil
}

func loadResponse(reader io.ReadCloser) (*hls_settings.SResponse, error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	resp := &hls_settings.SResponse{}
	if err := json.Unmarshal(body, resp); err != nil {
		return nil, err
	}

	if resp.FReturn != hls_settings.CErrorNone {
		return nil, fmt.Errorf("error code = %d", resp.FReturn)
	}

	return resp, nil
}

func deleteVoidStrings(s []string) []string {
	result := make([]string, 0, len(s))
	for _, v := range s {
		r := strings.TrimSpace(v)
		if r == "" {
			continue
		}
		result = append(result, r)
	}
	return result
}
