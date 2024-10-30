package common

import (
	"encoding/hex"
)

func NormalizeHex(s string) (string, error) {
	if len(s) >= 2 {
		if s[:2] == "0x" {
			s = s[2:]
		}
	}
	r, err := hex.DecodeString(s)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(r), nil
}
