package utils

import "golang.org/x/crypto/bcrypt"

const MinPasswordLen = 6

func GenerateHashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), MinPasswordLen)
	return string(bytes), err
}

func CompareHashPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
