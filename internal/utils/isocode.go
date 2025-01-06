package utils

var isoCodes = map[string]bool{
	"eng":     true, // English
	"swa":     true, // Swahili
	"default": true, // Default language: English
}

func IsValidISO639(code string) bool {
	return isoCodes[code]
}
