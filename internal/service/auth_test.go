package service

import (
	"errors"
	"testing"
	"time"

	"courier_managment_system_go/internal/domain"
	"courier_managment_system_go/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

func TestAuthServiceLogin(t *testing.T) {
	store := storage.NewMemoryStore()
	password := "secret123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	user, err := store.CreateUser(domain.User{
		Login:        "tester",
		PasswordHash: string(hash),
		Name:         "Tester",
		Role:         domain.RoleManager,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	service := NewAuthService(store, "secret", time.Hour)
	token, dto, err := service.Login(user.Login, password)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected token to be returned")
	}
	if dto.Login != user.Login {
		t.Fatalf("expected dto login %s, got %s", user.Login, dto.Login)
	}
}

func TestAuthServiceRejectsInvalidPassword(t *testing.T) {
	store := storage.NewMemoryStore()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	_, err := store.CreateUser(domain.User{
		Login:        "tester2",
		PasswordHash: string(hash),
		Name:         "Tester2",
		Role:         domain.RoleManager,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	service := NewAuthService(store, "secret", time.Hour)
	_, _, err = service.Login("tester2", "wrong")
	if err == nil {
		t.Fatal("expected error for invalid password")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
}
