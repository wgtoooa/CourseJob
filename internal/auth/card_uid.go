package authA

import (
	"regexp"
	"strings"
)

var uidHexRegex = regexp.MustCompile(`^[A-F0-9]+$`)

// Supported NFC UID lengths:
// - legacy: 4, 7
// - common ISO14443 card UIDs: 8, 14, 16, 20 (hex chars)
var supportedUIDLengths = map[int]struct{}{
	4:  {},
	7:  {},
	8:  {},
	14: {},
	16: {},
	20: {},
}

func NormalizeCardUID(uid string) string {
	normalized := strings.ToUpper(strings.TrimSpace(uid))
	normalized = strings.TrimPrefix(normalized, "0X")
	replacer := strings.NewReplacer(" ", "", ":", "", "-", "")
	return replacer.Replace(normalized)
}

func IsValidCardUID(uid string) bool {
	if uid == "" || !uidHexRegex.MatchString(uid) {
		return false
	}

	_, ok := supportedUIDLengths[len(uid)]
	return ok
}

func NormalizeAndValidateCardUID(uid string) (string, bool) {
	normalized := NormalizeCardUID(uid)
	return normalized, IsValidCardUID(normalized)
}
