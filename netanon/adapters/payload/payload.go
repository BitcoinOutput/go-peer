package payload

import "github.com/number571/go-peer/payload"

func NewPayload(head uint32, body []byte) IPayload {
	return payload.NewPayload(uint64(head), body)
}