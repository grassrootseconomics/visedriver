package common

import (
	"testing"
)

func TestParseAndScaleAmount(t *testing.T) {
	tests := []struct {
		name        string
		amount      string
		decimals    string
		want        string
		expectError bool
	}{
		{
			name:        "whole number",
			amount:      "123",
			decimals:    "2",
			want:        "12300",
			expectError: false,
		},
		{
			name:        "decimal number",
			amount:      "123.45",
			decimals:    "2",
			want:        "12345",
			expectError: false,
		},
		{
			name:        "zero decimals",
			amount:      "123.45",
			decimals:    "0",
			want:        "123",
			expectError: false,
		},
		{
			name:        "large number",
			amount:      "1000000.01",
			decimals:    "6",
			want:        "1000000010000",
			expectError: false,
		},
		{
			name:        "invalid amount",
			amount:      "abc",
			decimals:    "2",
			want:        "",
			expectError: true,
		},
		{
			name:        "invalid decimals",
			amount:      "123.45",
			decimals:    "abc",
			want:        "",
			expectError: true,
		},
		{
			name:        "zero amount",
			amount:      "0",
			decimals:    "2",
			want:        "0",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAndScaleAmount(tt.amount, tt.decimals)

			// Check error cases
			if tt.expectError {
				if err == nil {
					t.Errorf("ParseAndScaleAmount(%q, %q) expected error, got nil", tt.amount, tt.decimals)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseAndScaleAmount(%q, %q) unexpected error: %v", tt.amount, tt.decimals, err)
				return
			}

			if got != tt.want {
				t.Errorf("ParseAndScaleAmount(%q, %q) = %v, want %v", tt.amount, tt.decimals, got, tt.want)
			}
		})
	}
}
