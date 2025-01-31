package puzzle

import (
	"bytes"
	"math"
	"math/big"

	"github.com/number571/go-peer/pkg/crypto/hashing"
	"github.com/number571/go-peer/pkg/encoding"
)

var (
	_ IPuzzle = &sPoWPuzzle{}
)

type sPoWPuzzle struct {
	fDiff uint8
}

func NewPoWPuzzle(pDiff uint64) IPuzzle {
	return &sPoWPuzzle{
		fDiff: uint8(pDiff),
	}
}

// Proof of work by the method of finding the desired hash.
// Hash must start with 'diff' number of zero bits.
func (p *sPoWPuzzle) ProofBytes(packHash []byte) uint64 {
	var (
		target  = big.NewInt(1)
		intHash = big.NewInt(1)
		nonce   = uint64(0)
		hash    []byte
	)
	target.Lsh(target, hashSizeInBits()-uint(p.fDiff))
	for nonce < math.MaxUint64 {
		bNonce := encoding.Uint64ToBytes(nonce)
		hash = hashing.NewSHA256Hasher(bytes.Join(
			[][]byte{
				packHash,
				bNonce[:],
			},
			[]byte{},
		)).ToBytes()
		intHash.SetBytes(hash)
		if intHash.Cmp(target) == -1 {
			return nonce
		}
		nonce++
	}
	return nonce
}

// Verifies the work of the proof of work function.
func (p *sPoWPuzzle) VerifyBytes(packHash []byte, nonce uint64) bool {
	intHash := big.NewInt(1)
	target := big.NewInt(1)
	bNonce := encoding.Uint64ToBytes(nonce)
	hash := hashing.NewSHA256Hasher(bytes.Join(
		[][]byte{
			packHash,
			bNonce[:],
		},
		[]byte{},
	)).ToBytes()
	intHash.SetBytes(hash)
	target.Lsh(target, hashSizeInBits()-uint(p.fDiff))
	return intHash.Cmp(target) == -1
}

func hashSizeInBits() uint {
	return uint(hashing.CSHA256Size * 8)
}
