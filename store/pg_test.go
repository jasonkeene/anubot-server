package store_test

import (
	"crypto/rand"
	"testing"

	"golang.org/x/crypto/chacha20poly1305"

	"github.com/a8m/expect"
	"github.com/jasonkeene/anubot-server/store"
)

func TestBackendCreation(t *testing.T) {
	expect := expect.New(t)
	_, err := store.NewPostgres("", randomKey())
	expect(err).To.Be.Nil()
}

func TestBackendWillFailWhenKeyIsNotTheRightLength(t *testing.T) {
	expect := expect.New(t)
	_, err := store.NewPostgres("", randomBadKey())
	expect(err).Not.To.Be.Nil()
}

func randomBadKey() []byte {
	key := make([]byte, chacha20poly1305.KeySize-2)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	return key
}
