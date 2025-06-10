package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func CreatePassword() string {
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	return (string(hash))
}
