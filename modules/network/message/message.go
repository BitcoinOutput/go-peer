package message

import (
	"bytes"

	"github.com/number571/go-peer/modules/crypto/hashing"
	"github.com/number571/go-peer/modules/payload"
)

var (
	_ IMessage = &sMessage{}
)

type sMessage struct {
	fHash    []byte
	fPayload payload.IPayload
}

func NewMessage(pld payload.IPayload, key []byte) IMessage {
	return &sMessage{
		fHash: hashing.NewHMACSHA256Hasher(
			key,
			pld.ToBytes(),
		).Bytes(),
		fPayload: pld,
	}
}

func LoadMessage(packData, key []byte) IMessage {
	// check Hash[uN]
	if len(packData) < hashing.CSHA256Size {
		return nil
	}

	hashRecv := packData[:hashing.CSHA256Size]
	payloadBytes := packData[hashing.CSHA256Size:]
	if !bytes.Equal(
		hashRecv,
		hashing.NewHMACSHA256Hasher(
			key,
			payloadBytes,
		).Bytes(),
	) {
		return nil
	}

	// check Head[u64]
	pld := payload.LoadPayload(payloadBytes)
	if pld == nil {
		return nil
	}

	return &sMessage{
		fHash:    hashRecv,
		fPayload: pld,
	}
}

func (msg *sMessage) Hash() []byte {
	return msg.fHash
}

func (msg *sMessage) Payload() payload.IPayload {
	return msg.fPayload
}

func (msg *sMessage) ToBytes() []byte {
	return bytes.Join(
		[][]byte{
			msg.fHash,
			msg.fPayload.ToBytes(),
		},
		[]byte{},
	)
}
