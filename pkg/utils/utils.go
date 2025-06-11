package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
)

// GenerateRandomString creates a random string of the specified length.
func GenerateRandomString(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be greater than 0")
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes)[:length], nil
}

// FileExists checks if a file exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// CreateDirectoryIfNotExists creates a directory if it does not already exist.
func CreateDirectoryIfNotExists(path string) error {
	if !FileExists(path) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}