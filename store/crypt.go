package store

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/chacha20poly1305"
)

// Encrypt encrypts the plain text using ChaCha20-Poly1305.
func Encrypt(plain []byte, key []byte) (box string, err error) {
	nonce, err := generateNonce()
	if err != nil {
		return "", err
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return "", err
	}
	dst := make([]byte, 0, 1024)
	cypher := aead.Seal(dst, nonce, plain, nil)

	return encode(nonce, cypher), nil
}

// Decrypt decrypts the box string using ChaCha20-Poly1305.
func Decrypt(box string, key []byte) (plain []byte, err error) {
	nonce, cypher, err := decode(box)
	if err != nil {
		return nil, err
	}

	// nil in seal/open??
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}
	dst := make([]byte, 0, 1024)
	result, err := aead.Open(dst, nonce, cypher, nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func generateNonce() ([]byte, error) {
	nonce := make([]byte, chacha20poly1305.NonceSize)
	_, err := rand.Read(nonce)
	if err != nil {
		return nil, err
	}
	return nonce, nil
}

func encode(nonce, cypher []byte) (box string) {
	b64nonce := base64.RawStdEncoding.EncodeToString(nonce)
	b64cypher := base64.RawStdEncoding.EncodeToString(cypher)
	return strings.Join([]string{b64nonce, b64cypher}, "$")
}

func decode(box string) (nonce, cypher []byte, err error) {
	parts := strings.Split(box, "$")
	if len(parts) != 2 {
		return nil, nil, errors.New("invalid len of box parts")
	}
	nonce, err = base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, nil, err
	}
	cypher, err = base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, nil, err
	}
	return nonce, cypher, nil
}
