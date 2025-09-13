package login

import (
	"github.com/google/uuid"
)

func GenerateToken() (string, error) {
	token, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return token.String(), nil
}
