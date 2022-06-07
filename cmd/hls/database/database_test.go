package database

import (
	"os"
	"testing"

	"github.com/number571/go-peer/crypto"
)

const (
	tcPathDB = "test.db"
)

const (
	tcMessageTitle     = "test-title"
	tcMessageBody      = "test-body"
	tcMessageRawConcat = tcMessageTitle + tcMessageBody
)

var (
	tgHash = crypto.NewHasher([]byte("test-hash")).Bytes()
	tgDB   IKeyValueDB
)

func testHmsDefaultInit(path string) {
	tgDB = NewKeyValueDB(path)
}

func TestDB(t *testing.T) {
	testHmsDefaultInit(tcPathDB)
	defer os.RemoveAll(tcPathDB)

	err := tgDB.Push(tgHash)
	if err != nil {
		t.Error(err)
	}

	exist := tgDB.Exist(tgHash)
	if !exist {
		t.Errorf("load msg is nil")
	}

	err = tgDB.Close()
	if err != nil {
		t.Error(err)
	}
}
