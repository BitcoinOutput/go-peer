package block

import (
	"fmt"
	"testing"

	"github.com/number571/go-peer/cmd/ubc/kernel/transaction"
	"github.com/number571/go-peer/internal/testutils"
	"github.com/number571/go-peer/pkg/crypto/asymmetric"
	"github.com/number571/go-peer/pkg/crypto/hashing"
)

func TestTransaction(t *testing.T) {
	sett := NewSettings(&SSettings{})

	priv := asymmetric.LoadRSAPrivKey(testutils.TcPrivKey1024)
	hash := hashing.NewSHA256Hasher([]byte("prev-hash")).Bytes()

	txs := []transaction.ITransaction{}
	for i := uint64(0); i < sett.GetCountTXs(); i++ {
		tx := transaction.NewTransaction(
			sett.GetTransactionSettings(),
			priv,
			[]byte(fmt.Sprintf("transaction-%d", i)),
		)
		txs = append(txs, tx)
	}

	newBlock := NewBlock(sett, priv, hash, txs)
	if newBlock == nil {
		t.Errorf("new block is nil")
		return
	}

	if !newBlock.IsValid() {
		t.Errorf("new block is not valid")
		return
	}

	loadBlock := LoadBlock(sett, testutils.TcLargeBody)
	if loadBlock == nil {
		t.Errorf("load block is nil")
		return
	}

	if !loadBlock.IsValid() {
		t.Errorf("load block is not valid")
		return
	}
}
