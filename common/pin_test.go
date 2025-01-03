package common

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestIsValidPIN(t *testing.T) {
	tests := []struct {
		name     string
		pin      string
		expected bool
	}{
		{
			name:     "Valid PIN with 4 digits",
			pin:      "1234",
			expected: true,
		},
		{
			name:     "Valid PIN with leading zeros",
			pin:      "0001",
			expected: true,
		},
		{
			name:     "Invalid PIN with less than 4 digits",
			pin:      "123",
			expected: false,
		},
		{
			name:     "Invalid PIN with more than 4 digits",
			pin:      "12345",
			expected: false,
		},
		{
			name:     "Invalid PIN with letters",
			pin:      "abcd",
			expected: false,
		},
		{
			name:     "Invalid PIN with special characters",
			pin:      "12@#",
			expected: false,
		},
		{
			name:     "Empty PIN",
			pin:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsValidPIN(tt.pin)
			if actual != tt.expected {
				t.Errorf("IsValidPIN(%q) = %v; expected %v", tt.pin, actual, tt.expected)
			}
		})
	}
}

func TestHashPIN(t *testing.T) {
	tests := []struct {
		name string
		pin  string
	}{
		{
			name: "Valid PIN with 4 digits",
			pin:  "1234",
		},
		{
			name: "Valid PIN with leading zeros",
			pin:  "0001",
		},
		{
			name: "Empty PIN",
			pin:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPIN, err := HashPIN(tt.pin)
			if err != nil {
				t.Errorf("HashPIN(%q) returned an error: %v", tt.pin, err)
				return
			}

			if hashedPIN == "" {
				t.Errorf("HashPIN(%q) returned an empty hash", tt.pin)
			}

			// Ensure the hash can be verified with bcrypt
			err = bcrypt.CompareHashAndPassword([]byte(hashedPIN), []byte(tt.pin))
			if tt.pin != "" && err != nil {
				t.Errorf("HashPIN(%q) produced a hash that does not match: %v", tt.pin, err)
			}
		})
	}
}

func TestVerifyMigratedHashPin(t *testing.T) {
	tests := []struct {
		pin  string
		hash string
	}{
		{
			pin:  "1234",
			hash: "$2b$08$dTvIGxCCysJtdvrSnaLStuylPoOS/ZLYYkxvTeR5QmTFY3TSvPQC6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.pin, func(t *testing.T) {
			ok := VerifyPIN(tt.hash, tt.pin)
			if !ok {
				t.Errorf("VerifyPIN could not verify migrated PIN: %v", tt.pin, ok)
			}
		})
	}
}

func TestVerifyPIN(t *testing.T) {
	tests := []struct {
		name       string
		pin        string
		hashedPIN  string
		shouldPass bool
	}{
		{
			name:       "Valid PIN verification",
			pin:        "1234",
			hashedPIN:  hashPINHelper("1234"),
			shouldPass: true,
		},
		{
			name:       "Invalid PIN verification with incorrect PIN",
			pin:        "5678",
			hashedPIN:  hashPINHelper("1234"),
			shouldPass: false,
		},
		{
			name:       "Invalid PIN verification with empty PIN",
			pin:        "",
			hashedPIN:  hashPINHelper("1234"),
			shouldPass: false,
		},
		{
			name:       "Invalid PIN verification with invalid hash",
			pin:        "1234",
			hashedPIN:  "invalidhash",
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyPIN(tt.hashedPIN, tt.pin)
			if result != tt.shouldPass {
				t.Errorf("VerifyPIN(%q, %q) = %v; expected %v", tt.hashedPIN, tt.pin, result, tt.shouldPass)
			}
		})
	}
}

// Helper function to hash a PIN for testing purposes
func hashPINHelper(pin string) string {
	hashedPIN, err := HashPIN(pin)
	if err != nil {
		panic("Failed to hash PIN for test setup: " + err.Error())
	}
	return hashedPIN
}
