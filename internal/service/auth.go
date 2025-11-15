package service

import (
	"time"

	"courier_managment_system_go/internal/domain"
	"courier_managment_system_go/internal/storage"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	store    *storage.MemoryStore
	secret   []byte
	tokenTTL time.Duration
}

type Claims struct {
	UserID int64           `json:"user_id"`
	Login  string          `json:"login"`
	Role   domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(store *storage.MemoryStore, secret string, ttl time.Duration) *AuthService {
	return &AuthService{
		store:    store,
		secret:   []byte(secret),
		tokenTTL: ttl,
	}
}

func (s *AuthService) Login(login, password string) (string, UserDTO, error) {
	user, ok := s.store.GetUserByLogin(login)
	if !ok {
		return "", UserDTO{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", UserDTO{}, ErrInvalidCredentials
	}

	now := time.Now()
	claims := Claims{
		UserID: user.ID,
		Login:  user.Login,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Login,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", UserDTO{}, err
	}

	return signed, NewUserDTO(*user), nil
}

func (s *AuthService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidCredentials
	}
	return claims, nil
}

type AuthUser struct {
	ID    int64
	Login string
	Role  domain.UserRole
}

func UserFromClaims(c *Claims) AuthUser {
	return AuthUser{
		ID:    c.UserID,
		Login: c.Login,
		Role:  c.Role,
	}
}
