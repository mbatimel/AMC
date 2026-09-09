package service

import (
	"strings"
	"testing"
)

func TestGeneratePassword_LengthAndAlphabet(t *testing.T) {
	seen := map[byte]bool{}
	for i := 0; i < 200; i++ {
		password, err := generatePassword()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(password) != passwordLength {
			t.Fatalf("expected length %d, got %d (%q)", passwordLength, len(password), password)
		}
		for _, c := range []byte(password) {
			if !strings.ContainsRune(passwordAlphabet, rune(c)) {
				t.Fatalf("unexpected character %q in password %q", c, password)
			}
			seen[c] = true
		}
	}
	if len(seen) < 20 {
		t.Fatalf("expected reasonable character variety across 200 generations, saw only %d distinct chars", len(seen))
	}
}

func TestGeneratePassword_Unique(t *testing.T) {
	a, err := generatePassword()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := generatePassword()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == b {
		t.Fatalf("expected two generated passwords to differ, got the same value twice: %q", a)
	}
}
