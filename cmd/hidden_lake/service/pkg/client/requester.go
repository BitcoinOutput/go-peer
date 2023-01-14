package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	pkg_settings "github.com/number571/go-peer/cmd/hidden_lake/service/pkg/settings"
	"github.com/number571/go-peer/pkg/crypto/asymmetric"
	"github.com/number571/go-peer/pkg/encoding"
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

func (requester *sRequester) Request(push *pkg_settings.SPush) ([]byte, error) {
	res, err := doRequest(
		http.MethodPost,
		requester.fHost+pkg_settings.CHandleNetworkPush,
		push,
	)
	if err != nil {
		return nil, err
	}

	return encoding.HexDecode(res), nil
}

func (requester *sRequester) Broadcast(push *pkg_settings.SPush) error {
	_, err := doRequest(
		http.MethodPut,
		requester.fHost+pkg_settings.CHandleNetworkPush,
		push,
	)
	return err
}

func (requester *sRequester) GetFriends() (map[string]asymmetric.IPubKey, error) {
	res, err := doRequest(
		http.MethodGet,
		requester.fHost+pkg_settings.CHandleConfigFriends,
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

func (requester *sRequester) AddFriend(friend *pkg_settings.SFriend) error {
	_, err := doRequest(
		http.MethodPost,
		requester.fHost+pkg_settings.CHandleConfigFriends,
		friend,
	)
	return err
}

func (requester *sRequester) DelFriend(friend *pkg_settings.SFriend) error {
	_, err := doRequest(
		http.MethodDelete,
		requester.fHost+pkg_settings.CHandleConfigFriends,
		friend,
	)
	return err
}

func (requester *sRequester) GetOnlines() ([]string, error) {
	res, err := doRequest(
		http.MethodGet,
		requester.fHost+pkg_settings.CHandleNetworkOnline,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return deleteVoidStrings(strings.Split(res, ",")), nil
}

func (requester *sRequester) DelOnline(connect *pkg_settings.SConnect) error {
	_, err := doRequest(
		http.MethodDelete,
		requester.fHost+pkg_settings.CHandleNetworkOnline,
		connect,
	)
	return err
}

func (requester *sRequester) GetConnections() ([]string, error) {
	res, err := doRequest(
		http.MethodGet,
		requester.fHost+pkg_settings.CHandleConfigConnects,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return deleteVoidStrings(strings.Split(res, ",")), nil
}

func (requester *sRequester) AddConnection(connect *pkg_settings.SConnect) error {
	_, err := doRequest(
		http.MethodPost,
		requester.fHost+pkg_settings.CHandleConfigConnects,
		connect,
	)
	return err
}

func (requester *sRequester) DelConnection(connect *pkg_settings.SConnect) error {
	_, err := doRequest(
		http.MethodDelete,
		requester.fHost+pkg_settings.CHandleConfigConnects,
		connect,
	)
	return err
}

func (requester *sRequester) PrivKey(privKey *pkg_settings.SPrivKey) error {
	_, err := doRequest(
		http.MethodPost,
		requester.fHost+pkg_settings.CHandleNodeKey,
		privKey,
	)
	return err
}

func (requester *sRequester) PubKey() (asymmetric.IPubKey, error) {
	res, err := doRequest(
		http.MethodGet,
		requester.fHost+pkg_settings.CHandleNodeKey,
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

	req.Header.Set("Content-Type", pkg_settings.CContentType)
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

func loadResponse(reader io.ReadCloser) (*pkg_settings.SResponse, error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	resp := &pkg_settings.SResponse{}
	if err := json.Unmarshal(body, resp); err != nil {
		return nil, err
	}

	if resp.FReturn != pkg_settings.CErrorNone {
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