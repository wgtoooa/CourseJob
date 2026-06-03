package authA

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidHash         = errors.New("invalid hash format")
	ErrIncompatibleVersion = errors.New("incompatible hash version")
)

const (
	uidPrefix  = "$uidhmac"
	uidVersion = 1
)

type Hasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) (bool, error)
	NeedsRehash(encodedHash string) (bool, error)
}

type hasher struct {
	key []byte
}

func NewHasher(secret string) Hasher {
	return &hasher{key: []byte(secret)}
}

func (h *hasher) Hash(password string) (string, error) {
	digest := h.hmacSHA256([]byte(password))
	return fmt.Sprintf("%s$v=%d$%s", uidPrefix, uidVersion, hex.EncodeToString(digest)), nil
}

func (h *hasher) Verify(password string, encodedHash string) (bool, error) {
	stored, err := parseUIDHash(encodedHash)
	if err != nil {
		return false, err
	}

	computed := h.hmacSHA256([]byte(password))
	return subtle.ConstantTimeCompare(stored, computed) == 1, nil
}

func (h *hasher) NeedsRehash(encodedHash string) (bool, error) {
	_, err := parseUIDHash(encodedHash)
	if err != nil {
		return false, err
	}
	return false, nil
}

func (h *hasher) hmacSHA256(msg []byte) []byte {
	mac := hmac.New(sha256.New, h.key)
	mac.Write(msg)
	return mac.Sum(nil)
}

func parseUIDHash(encodedHash string) ([]byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 4 || parts[1] != "uidhmac" {
		return nil, ErrInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, ErrInvalidHash
	}
	if version != uidVersion {
		return nil, ErrIncompatibleVersion
	}

	decoded, err := hex.DecodeString(parts[3])
	if err != nil {
		return nil, ErrInvalidHash
	}
	return decoded, nil
}
