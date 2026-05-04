package utils

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/google/uuid"
)

// GenerateUUID generates a new UUID v4 string
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateNodeID generates a unique node ID for IoT devices
func GenerateNodeID() string {
	return uuid.New().String()
}

// GenerateAPIKey generates a cryptographically secure 32-byte hex API key
func GenerateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
