package store_test

import (
	"crypto/rand"
	"testing"

	"golang.org/x/crypto/chacha20poly1305"

	"github.com/a8m/expect"
	"github.com/jasonkeene/anubot-server/store"
)

func TestItEncryptsAndDecrypts(t *testing.T) {
	expect := expect.New(t)
	key := randomKey()
	plain := []byte("test-plain-text")
	box, err := store.Encrypt(plain, key)
	expect(err).To.Be.Nil()
	expect(box).Not.To.Equal(string(plain))
	result, err := store.Decrypt(box, key)
	expect(err).To.Be.Nil()
	expect(result).To.Equal(plain)
}

func randomKey() []byte {
	key := make([]byte, chacha20poly1305.KeySize)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	return key
}
