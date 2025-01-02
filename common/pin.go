package common

import "regexp"

// Define the regex pattern as a constant
const (
	pinPattern = `^\d{4}$`
)

// checks whether the given input is a 4 digit number
func IsValidPIN(pin string) bool {
	match, _ := regexp.MatchString(pinPattern, pin)
	return match
}
