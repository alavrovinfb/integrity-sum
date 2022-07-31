package hasher

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
)

//go:generate mockgen -source=hasher.go -destination=mocks/mock_hasher.go

// NewHashSum takes a hashing algorithm as input and returns a hash sum with other data or an error
func NewHashSum(alg string) (h hash.Hash, err error) {
	switch alg {
	case "MD5":
		h = md5.New()
	case "SHA1":
		h = sha1.New()
	case "SHA224":
		h = sha256.New224()
	case "SHA384":
		h = sha512.New384()
	case "SHA512":
		h = sha512.New()
	default:
		h = sha256.New()
	}
	return h, nil
}
