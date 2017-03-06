package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

func main() {
	key := make([]byte, chacha20poly1305.KeySize)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	encoded := base64.RawStdEncoding.EncodeToString(key)
	fmt.Printf("ANUBOT_ENCRYPTION_KEY=%s\n", encoded)
}
