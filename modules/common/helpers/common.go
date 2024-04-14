package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"strings"
)

// GetEnv - Lookup the environment variable provided and set to default value if variable isn't found
func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); len(value) > 0 {
		return value
	}

	return fallback
}

// GetResourceFromPath extracts the resource from the URL path /api/v1/<action>.
// Ignores potential subresources.
func GetResourceFromPath(path string) *string {
	if !strings.HasPrefix(path, "/api/v1") {
		return nil
	}

	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return nil
	}

	return &parts[3]
}

func RandomBytes(size int) []byte {
	bytes := make([]byte, size)
	_, _ = rand.Read(bytes)

	return bytes
}

func Random64BaseEncodedBytes(size int) string {
	bytes := RandomBytes(size)
	return base64.StdEncoding.EncodeToString(bytes)
}
