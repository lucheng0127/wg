package core

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/curve25519"
)

type Key struct {
	Priv string
	Pub  string
}

func NewRandomKey() (*Key, error) {
	// New random key
	privateKey := make([]byte, curve25519.ScalarSize)
	if _, err := rand.Read(privateKey); err != nil {
		return nil, err
	}

	// Formate key
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	// Generate key pairs
	publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, err
	}

	// Base64 encode
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKey)
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKey)

	return &Key{Priv: privateKeyBase64, Pub: publicKeyBase64}, nil
}
