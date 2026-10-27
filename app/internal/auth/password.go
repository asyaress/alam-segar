package auth

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func ValidatePassword(password string) error {
	if len(password) < 12 {
		return errors.New("kata sandi minimal 12 karakter")
	}
	var upper, lower, digit, symbol bool
	for _, value := range password {
		switch {
		case unicode.IsUpper(value):
			upper = true
		case unicode.IsLower(value):
			lower = true
		case unicode.IsDigit(value):
			digit = true
		default:
			symbol = true
		}
	}
	if !upper || !lower || !digit || !symbol {
		return errors.New("kata sandi harus memuat huruf besar, huruf kecil, angka, dan simbol")
	}
	return nil
}

func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
