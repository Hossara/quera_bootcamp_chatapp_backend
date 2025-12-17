package auth

import (
	"fmt"
	"time"

	"github.com/o1egl/paseto"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	pasetoKey       []byte
	tokenExpiration time.Duration
}

type TokenPayload struct {
	UserID   int       `json:"user_id"`
	Username string    `json:"username"`
	IssuedAt time.Time `json:"issued_at"`
	ExpireAt time.Time `json:"expire_at"`
}

func NewAuthService(pasetoKey string, expirationHours int) *AuthService {
	return &AuthService{
		pasetoKey:       []byte(pasetoKey),
		tokenExpiration: time.Duration(expirationHours) * time.Hour,
	}
}

func (s *AuthService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

func (s *AuthService) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (s *AuthService) CreateToken(userID int, username string) (string, error) {
	now := time.Now()
	payload := TokenPayload{
		UserID:   userID,
		Username: username,
		IssuedAt: now,
		ExpireAt: now.Add(s.tokenExpiration),
	}

	v2 := paseto.NewV2()
	token, err := v2.Encrypt(s.pasetoKey, payload, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
	}

	return token, nil
}

func (s *AuthService) VerifyToken(token string) (*TokenPayload, error) {
	v2 := paseto.NewV2()
	var payload TokenPayload

	err := v2.Decrypt(token, s.pasetoKey, &payload, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	if time.Now().After(payload.ExpireAt) {
		return nil, fmt.Errorf("token has expired")
	}

	return &payload, nil
}
