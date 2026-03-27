package validator

import "regexp"

func validUID(uid string) bool {
	var cardUIDRegex = regexp.MustCompile(`^[A-F0-9]{4,7}$`)
	return cardUIDRegex.MatchString(uid)
}
