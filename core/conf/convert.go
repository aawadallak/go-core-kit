package conf

import (
	"strconv"
)

// convertToInt converts a string value to an integer.
// Returns 0 if the conversion fails.
func convertToInt(value string) int {
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return v
}

// convertToBool converts a string value to a boolean.
// Returns false if the conversion fails.
func convertToBool(value string) bool {
	v, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return v
}
