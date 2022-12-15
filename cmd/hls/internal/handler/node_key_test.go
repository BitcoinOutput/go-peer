package handler

import (
	"fmt"
	"testing"

	"github.com/number571/go-peer/cmd/hls/pkg/client"
	"github.com/number571/go-peer/internal/testutils"
	"github.com/number571/go-peer/pkg/crypto/asymmetric"
)

func TestHandlePubKeyAPI(t *testing.T) {
	_, node, srv := testAllCreate(tcPathConfig, tcPathDB, testutils.TgAddrs[8])
	defer testAllFree(node, srv)

	client := client.NewClient(
		client.NewRequester(fmt.Sprintf("http://%s", testutils.TgAddrs[8])),
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

	privKey := asymmetric.LoadRSAPrivKey(testutils.TcPrivKey2)
	if err := client.PrivKey(privKey); err != nil {
		t.Errorf("failed update private key")
		return
	}

	newPubKey, err := client.PubKey()
	if err != nil {
		t.Error(err)
		return
	}

	if pubKey.Address().String() == newPubKey.Address().String() {
		t.Errorf("public keys are equals")
		return
	}
}