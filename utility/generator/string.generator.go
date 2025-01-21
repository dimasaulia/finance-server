package generator

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateRandomBase64Url() (string, error) {
	// Generate 32 random bytes
	randomValues := make([]byte, 32)
	_, err := rand.Read(randomValues)
	if err != nil {
		return "", err
	}

	// Encode to Base64 URL without padding
	base64Url := base64.URLEncoding.EncodeToString(randomValues)
	// Remove padding '='
	base64Url = base64Url[:len(base64Url)-len(base64Url)%4]

	return base64Url, nil
}
