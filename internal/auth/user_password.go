package authA

import "golang.org/x/crypto/bcrypt"

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) bool
}

type bcryptHasher struct {
	cost int
}

func NewPasswordHasher() PasswordHasher {
	return &bcryptHasher{cost: bcrypt.DefaultCost}
}

func (b *bcryptHasher) Hash(password string) (string, error) {
	digest, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", err
	}
	return string(digest), nil
}

func (b *bcryptHasher) Verify(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
