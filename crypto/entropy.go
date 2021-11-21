package crypto

import "bytes"

// Increase entropy by multiple hashing.
func RaiseEntropy(info, salt []byte, bits uint) []byte {
	lim := uint64(1 << bits)
	for i := uint64(0); i < lim; i++ {
		info = SumHash(bytes.Join(
			[][]byte{
				info,
				salt,
			},
			[]byte{},
		))
	}
	return info
}
