package message

import (
	"bytes"
	"testing"

	"github.com/number571/go-peer/pkg/payload"
)

const (
	tcHead = 12345
	tcBody = "hello, world!"
	tcKey  = "network-key"
)

func TestMessage(t *testing.T) {
	msg := NewMessage(
		payload.NewPayload(tcHead, []byte(tcBody)),
		[]byte(tcKey),
	)

	if !bytes.Equal(msg.GetPayload().GetBody(), []byte(tcBody)) {
		t.Error("payload body not equal body in message")
		return
	}

	if msg.GetPayload().GetHead() != tcHead {
		t.Error("payload head not equal head in message")
		return
	}

	msg1 := LoadMessage(msg.GetBytes(), []byte(tcKey))
	if !bytes.Equal(msg.GetPayload().ToBytes(), msg1.GetPayload().ToBytes()) {
		t.Error("load message not equal new message")
		return
	}
}
