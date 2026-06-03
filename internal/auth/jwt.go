package authA

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UserID    int64  `json:"uid"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenKind string `json:"token_kind"`
	jwt.RegisteredClaims
}

type TokenManager interface {
	GenerateAccess(userID int64, email string, role string) (token string, expiresAt time.Time, err error)
	GenerateRefresh(userID int64, email string, role string) (token string, tokenID string, expiresAt time.Time, err error)
	ParseAccess(token string) (*Claims, error)
	ParseRefresh(token string) (*Claims, error)
	AccessTTL() time.Duration
	RefreshTTL() time.Duration
}

type jwtManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	issuer     string
}

func NewTokenManager(secret string, accessTTL, refreshTTL time.Duration) TokenManager {
	return &jwtManager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		issuer:     "coursejob-api",
	}
}

func (j *jwtManager) AccessTTL() time.Duration {
	return j.accessTTL
}

func (j *jwtManager) RefreshTTL() time.Duration {
	return j.refreshTTL
}

func (j *jwtManager) GenerateAccess(userID int64, email string, role string) (string, time.Time, error) {
	token, _, expiresAt, err := j.generateToken(userID, email, role, "access", j.accessTTL, false)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

func (j *jwtManager) GenerateRefresh(userID int64, email string, role string) (string, string, time.Time, error) {
	return j.generateToken(userID, email, role, "refresh", j.refreshTTL, true)
}

func (j *jwtManager) generateToken(
	userID int64,
	email string,
	role string,
	tokenKind string,
	ttl time.Duration,
	withTokenID bool,
) (string, string, time.Time, error) {
	now := time.Now().UTC()
	exp := now.Add(ttl)

	tokenID := ""
	if withTokenID {
		var err error
		tokenID, err = generateTokenID()
		if err != nil {
			return "", "", time.Time{}, err
		}
	}

	claims := Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenKind: tokenKind,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(j.secret)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return signed, tokenID, exp, nil
}

func (j *jwtManager) ParseAccess(token string) (*Claims, error) {
	return j.parseToken(token, "access")
}

func (j *jwtManager) ParseRefresh(token string) (*Claims, error) {
	return j.parseToken(token, "refresh")
}

func (j *jwtManager) parseToken(token string, expectedKind string) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return j.secret, nil
	}, jwt.WithIssuer(j.issuer), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if !parsed.Valid {
		return nil, ErrInvalidToken
	}
	if claims.UserID == 0 || claims.Role == "" || claims.Email == "" {
		return nil, ErrInvalidToken
	}
	if claims.TokenKind == "" && expectedKind == "access" {
		// Backward compatibility for previously issued access tokens without token_kind.
		return claims, nil
	}
	if claims.TokenKind != expectedKind {
		return nil, ErrInvalidToken
	}
	if expectedKind == "refresh" && claims.ID == "" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func generateTokenID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
