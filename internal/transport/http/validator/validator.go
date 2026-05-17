package validator

import (
	"net/mail"
	"regexp"
)

func validUID(uid string) bool {
	var cardUIDRegex = regexp.MustCompile(`^[A-F0-9]{4,7}$`)
	return cardUIDRegex.MatchString(uid)
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
