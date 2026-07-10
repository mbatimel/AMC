package repo

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const CodeLength = 6

// codeUpperBound = 10^CodeLength, keeps generated codes zero-padded to CodeLength digits.
var codeUpperBound = big.NewInt(1000000)

func GenerateVerificationCode() (int64, error) {
	num, err := rand.Int(rand.Reader, codeUpperBound)
	if err != nil {
		return 0, err
	}
	return num.Int64(), nil
}
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func NormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email must not be empty")
	}

	if len(email) < 3 || len(email) > 254 {
		return fmt.Errorf("email length must be between 3 and 254 characters")
	}

	if strings.Count(email, "@") != 1 {
		return fmt.Errorf("email must contain exactly one @")
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format: %w", err)
	}

	return nil
}

func ValidateDateString(dateStr string) (string, error) {
	if dateStr == "" {
		return "", nil
	}

	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		return "", fmt.Errorf("date must be in YYYY-MM-DD format: %w", err)
	}

	return dateStr, nil
}
