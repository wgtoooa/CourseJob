package validator

import (
	authA "CourseJob/internal/auth"
	"net/mail"
)

func validUID(uid string) bool {
	_, ok := authA.NormalizeAndValidateCardUID(uid)
	return ok
}

func validEmail(email string) bool {
	if email == "" {
		return false
	}
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	return addr.Address == email
}
