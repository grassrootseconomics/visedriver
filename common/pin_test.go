package common

import "testing"

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
