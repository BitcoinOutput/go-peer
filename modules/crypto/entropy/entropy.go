package entropy

import (
	"bytes"

	"github.com/number571/go-peer/modules/crypto/hashing"
)

var (
	_ IEntropy = &sEntropy{}
)

type sEntropy struct {
	bits uint64
}

func NewEntropy(bits uint64) IEntropy {
	return &sEntropy{
		bits: bits,
	}
}

// Increase entropy by multiple hashing.
func (entr *sEntropy) Raise(data, salt []byte) []byte {
	lim := uint64(1 << entr.bits)

	for i := uint64(0); i < lim; i++ {
		data = hashing.NewSHA256Hasher(bytes.Join(
			[][]byte{
				data,
				salt,
			},
			[]byte{},
		)).Bytes()
	}

	return data
}