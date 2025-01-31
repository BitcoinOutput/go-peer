package hashing

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

var (
	_ IHasher = &sHMACSHA256Hasher{}
)

type sHMACSHA256Hasher struct {
	fHash []byte
}

func NewHMACSHA256Hasher(pKey []byte, pData []byte) IHasher {
	h := hmac.New(sha256.New, pKey)
	h.Write(pData)
	return &sHMACSHA256Hasher{
		fHash: h.Sum(nil),
	}
}

func (p *sHMACSHA256Hasher) ToString() string {
	return fmt.Sprintf("HMAC(%s){%X}", p.GetType(), p.ToBytes())
}

func (p *sHMACSHA256Hasher) ToBytes() []byte {
	return p.fHash
}

func (p *sHMACSHA256Hasher) GetType() string {
	return CSHA256KeyType
}

func (p *sHMACSHA256Hasher) GetSize() uint64 {
	return CSHA256Size
}
