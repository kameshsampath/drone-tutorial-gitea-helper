package utils

import (
	"crypto/rand"
	"encoding/hex"
)

//RandomHex generates and returns a random 16 digit Hex value
func RandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
