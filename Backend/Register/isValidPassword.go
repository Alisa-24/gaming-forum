package register

import (
	"errors"
	"strings"
)

func isValidPassword(password string) (bool, error) {
	if len(password) < 8 {
		return false, errors.New("password must be at least 8 characters long")
	} // Password too short
	if len(password) > 24 {
		return false, errors.New("password must not exceed 24 characters")
	} // Password too long
	if strings.Contains(password, " ") {
		return false, errors.New("password must not contain spaces")
	}
	return true, nil // Password is valid
}
