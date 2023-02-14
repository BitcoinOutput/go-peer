package app

import (
	"fmt"
	"os"
	"testing"

	"github.com/number571/go-peer/cmd/hidden_lake/service/internal/config"
	"github.com/number571/go-peer/cmd/hidden_lake/service/pkg/client"
	testutils "github.com/number571/go-peer/test/_data"

	hls_settings "github.com/number571/go-peer/cmd/hidden_lake/service/internal/settings"
	anon_testutils "github.com/number571/go-peer/test/_data/anonymity"
)

const (
	tcPathDB     = hls_settings.CPathDB
	tcPathConfig = hls_settings.CPathCFG
)

func testDeleteFiles() {
	os.RemoveAll(tcPathDB)
	os.RemoveAll(tcPathConfig)
}

func TestApp(t *testing.T) {
	testDeleteFiles()
	defer testDeleteFiles()

	// Run application
	cfg, err := config.NewConfig(tcPathConfig, &config.SConfig{
		FAddress: &config.SAddress{
			FTCP:  testutils.TgAddrs[14],
			FHTTP: testutils.TgAddrs[15],
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	node := anon_testutils.TestNewNode(tcPathDB)
	app := NewApp(cfg, node)
	if err := app.Run(); err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := app.Close(); err != nil {
			t.Error(err)
			return
		}
	}()

	client := client.NewClient(
		client.NewBuilder(),
		client.NewRequester(fmt.Sprintf("http://%s", testutils.TgAddrs[15])),
	)

	// Check public key of node
	pubKey, err := client.GetPubKey()
	if err != nil {
		t.Error(err)
		return
	}

	if pubKey.String() != node.Queue().Client().PubKey().String() {
		t.Errorf("public keys are not equals")
		return
	}
}
