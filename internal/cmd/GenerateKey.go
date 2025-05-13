package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

type GenerateKey struct{}

func (cmd *GenerateKey) Run() error {
	key := make([]byte, 30)
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("error while generating a random key: %w", err)
	}
	fmt.Println(strings.ToUpper(hex.EncodeToString(key)))
	return nil
}
