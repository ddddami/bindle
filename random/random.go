package random

import (
	"crypto/rand"
	"fmt"
)

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Options struct {
	Length int
}

// Generate creates a random string of the specified length using the characters defined in the chars constant.
func Generate(opts Options) (string, error) {
	if opts.Length < 0 {
		return "", fmt.Errorf("length must be greater than -1")
	}

	bytes := make([]byte, opts.Length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("error generating random bytes: %w", err)
	}

	charLen := len(chars)
	for i, b := range bytes {
		bytes[i] = chars[b%byte(charLen)]
	}

	return string(bytes), nil
}
