package util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(s string) (string, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}

	return string(hashedPass), nil
}

func ComparePassword(s string, hashedPass string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(s))
}