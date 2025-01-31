package payload

import (
	"bytes"
	"testing"

	testutils "github.com/number571/go-peer/test/_data"
)

func TestPayload(t *testing.T) {
	pl := NewPayload(uint64(testutils.TcHead), []byte(testutils.TcBody))

	decPl := LoadPayload(pl.ToBytes())
	if decPl == nil {
		t.Error("decode payload is nil")
		return
	}

	if !bytes.Equal(pl.GetBody(), decPl.GetBody()) {
		t.Error("data not equal with decoded version of payload")
		return
	}

	if pl.GetHead() != decPl.GetHead() {
		t.Error("title not equal with decoded version of payload")
		return
	}
}
