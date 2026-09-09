package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const passwordLength = 16
const passwordAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// generatePassword returns a cryptographically random password of
// passwordLength characters drawn uniformly from passwordAlphabet (no
// modulo bias — each character comes from its own rand.Int draw).
func generatePassword() (string, error) {
	alphabetLen := big.NewInt(int64(len(passwordAlphabet)))
	result := make([]byte, passwordLength)
	for i := range result {
		n, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			return "", fmt.Errorf("generate password: %w", err)
		}
		result[i] = passwordAlphabet[n.Int64()]
	}
	return string(result), nil
}
