package service

import (
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"

	customErrors "github.com/mbatimel/AMC/auth/internal/errors"
)

const (
	maxEmailLength = 255
	maxPhoneLength = 255
)

func canonicalEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeRegistrationEmail(value string) (string, error) {
	value = canonicalEmail(value)
	if value == "" || utf8.RuneCountInString(value) > maxEmailLength || strings.ContainsAny(value, "\r\n\x00") {
		return "", customErrors.ValidationError("email")
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return "", customErrors.ValidationError("email")
	}
	return value, nil
}

// normalizeRegistrationPhone intentionally follows the users service's
// existing phone rules. It trims surrounding whitespace but preserves the
// accepted display format that is stored in users.phone.
func normalizeRegistrationPhone(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", customErrors.ValidationError("phone")
	}
	digits := 0
	for _, r := range value {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case strings.ContainsRune("+()- ", r):
		default:
			return "", customErrors.ValidationError("phone")
		}
	}
	if digits < 7 || digits > 15 || utf8.RuneCountInString(value) > maxPhoneLength {
		return "", customErrors.ValidationError("phone")
	}
	return value, nil
}

// normalizeINN returns the only representation used by the rest of the
// registration flow: digits without whitespace or hyphens.
func normalizeINN(value string) (string, error) {
	value = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '-' {
			return -1
		}
		return r
	}, value)
	if value == "" {
		return "", customErrors.InnEmptyErr("inn")
	}
	if len(value) != 10 && len(value) != 12 {
		return "", customErrors.InnInvalidError(value)
	}

	digits := make([]int, len(value))
	for i, ch := range value {
		if ch < '0' || ch > '9' {
			return "", customErrors.InnInvalidError(value)
		}
		digits[i] = int(ch - '0')
	}

	valid := validate10(digits)
	if len(digits) == 12 {
		valid = validate12(digits)
	}
	if !valid {
		return "", customErrors.InnInvalidError(value)
	}
	return value, nil
}

func validate10(d []int) bool {
	coeffs := []int{2, 4, 10, 3, 5, 9, 4, 6, 8}
	sum := 0
	for i := 0; i < 9; i++ {
		sum += d[i] * coeffs[i]
	}
	control := (sum % 11) % 10
	return control == d[9]
}

func validate12(d []int) bool {
	coeffs11 := []int{7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
	sum11 := 0
	for i := 0; i < 10; i++ {
		sum11 += d[i] * coeffs11[i]
	}
	control11 := (sum11 % 11) % 10
	if control11 != d[10] {
		return false
	}

	coeffs12 := []int{3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
	sum12 := 0
	for i := 0; i < 11; i++ {
		sum12 += d[i] * coeffs12[i]
	}
	control12 := (sum12 % 11) % 10
	return control12 == d[11]
}
