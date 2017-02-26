package store_test

import (
	"testing"

	"github.com/a8m/expect"
	"github.com/jasonkeene/anubot-server/store"
)

func TestHashVerification(t *testing.T) {
	expect := expect.New(t)
	password := "test-password"
	hash, err := store.Hash(password)
	expect(err).To.Be.Nil()
	valid, err := store.Verify(password, hash)
	expect(err).To.Be.Nil()
	expect(valid).To.Be.True()
	valid, err = store.Verify("bad-password", hash)
	expect(err).To.Be.Nil()
	expect(valid).To.Be.False()
}
