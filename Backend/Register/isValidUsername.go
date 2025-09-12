package register

import (
	"errors"
	"strings"
	"unicode"
)

func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 20 {
		return errors.New("username must be between 3 and 20 characters")
	}
	if strings.Contains(username, " ") {
		return errors.New("username must not contain spaces")
	}
	if strings.ContainsAny(username, "!@#$%^&*()+={}[]|\\:;\"'<>,?/-") {
		return errors.New("username must not contain special characters")
	}
	if strings.Contains(username, "..") {
		return errors.New("username must not contain consecutive dots")
	}
	if strings.HasPrefix(username, ".") || strings.HasSuffix(username, ".") {
		return errors.New("username must not start or end with a dot")
	}
	if strings.HasPrefix(username, "_") || strings.HasSuffix(username, "_") {
		return errors.New("username must not start or end with an underscore")
	}

	hasAlphaNum := false
	for _, r := range username {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			hasAlphaNum = true
			break
		}
	}
	if !hasAlphaNum {
		return errors.New("username must contain at least one alphanumeric character")
	}

	return nil // valid username
}
