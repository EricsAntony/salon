package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashString creates a SHA-256 hash of the input string
func HashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// HashBytes creates a SHA-256 hash of the input bytes
func HashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
