package login

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateToken() (string, error) {
	byteToken := make([]byte, 32) // Generate a 32-byte token
	_, err := rand.Read(byteToken)
	if err != nil {
		return "", err // Return error if token generation fails
	}
	return hex.EncodeToString(byteToken), nil // Convert byte slice to hex string
}
