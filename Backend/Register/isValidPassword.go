package register

import (
	"errors"
	"strings"
)

const specialChars = "!@#$%^&*-_"

func isValidPassword(password string) (bool, error) {
	if len(password) < 8 {
		return false, errors.New("password must be at least 8 characters long")
	} // Password too short
	if len(password) > 64 {
		return false, errors.New("password must not exceed 64 characters")
	} // Password too long
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true // Contains uppercase letter
		} else if char >= 'a' && char <= 'z' {
			hasLower = true // Contains lowercase letter
		} else if char >= '0' && char <= '9' {
			hasDigit = true // Contains digit
		} else if strings.ContainsRune(specialChars, char) {
			hasSpecial = true // Contains special character
		}
		if strings.Contains(password, " ") {
			return false, errors.New("password must not contain spaces")
		}
	}
	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return false, errors.New("password must contain at least one uppercase letter, one lowercase letter, one digit, and one special character")
	}
	return true, nil // Password is valid
}
