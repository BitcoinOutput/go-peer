package handler

import (
	"fmt"
	"testing"

	"github.com/number571/go-peer/cmd/hls/hlc"
	"github.com/number571/go-peer/settings/testutils"
)

func TestHandlePubKeyAPI(t *testing.T) {
	_, node, srv := testAllCreate(tcPathConfig, tcPathDB, testutils.TgAddrs[8])
	defer testAllFree(node, srv)

	client := hlc.NewClient(
		hlc.NewRequester(fmt.Sprintf("http://%s", testutils.TgAddrs[8])),
	)

	pubKey, err := client.PubKey()
	if err != nil {
		t.Error(err)
		return
	}

	if pubKey.String() != node.Queue().Client().PubKey().String() {
		t.Errorf("public keys not equals")
		return
	}
}
