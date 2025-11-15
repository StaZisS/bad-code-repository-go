package service

import (
	"strings"
	"time"

	"courier_managment_system_go/internal/domain"
	"courier_managment_system_go/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	store        *storage.MemoryStore
	passwordCost int
}

func NewUserService(store *storage.MemoryStore) *UserService {
	return &UserService{
		store:        store,
		passwordCost: bcrypt.DefaultCost,
	}
}

type CreateUserInput struct {
	Login    string
	Password string
	Name     string
	Role     domain.UserRole
}

type UpdateUserInput struct {
	Name     *string
	Login    *string
	Role     *domain.UserRole
	Password *string
}

func (s *UserService) EnsureAdminUser(login, password, name string) error {
	login = strings.TrimSpace(login)
	if login == "" {
		login = "admin"
	}
	if name == "" {
		name = "Системный администратор"
	}
	if password == "" {
		password = "admin123"
	}
	if _, ok := s.store.GetUserByLogin(login); ok {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.passwordCost)
	if err != nil {
		return err
	}
	_, err = s.store.CreateUser(domain.User{
		Login:        login,
		PasswordHash: string(hash),
		Name:         name,
		Role:         domain.RoleAdmin,
		CreatedAt:    time.Now().UTC(),
	})
	return err
}

func (s *UserService) ListUsers(role *domain.UserRole) []UserDTO {
	users := s.store.ListUsers(role)
	dtos := make([]UserDTO, len(users))
	for i, user := range users {
		dtos[i] = NewUserDTO(user)
	}
	return dtos
}

func (s *UserService) CreateUser(input CreateUserInput) (UserDTO, error) {
	if errs := validateUserInput(input); len(errs) > 0 {
		return UserDTO{}, ValidationErrors(errs)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), s.passwordCost)
	if err != nil {
		return UserDTO{}, err
	}
	user, err := s.store.CreateUser(domain.User{
		Login:        strings.TrimSpace(input.Login),
		PasswordHash: string(hash),
		Name:         strings.TrimSpace(input.Name),
		Role:         input.Role,
		CreatedAt:    time.Now().UTC(),
	})
	if err != nil {
		if err == storage.ErrConflict {
			return UserDTO{}, ValidationErrors{{
				Field:   "login",
				Message: "user with this login already exists",
			}}
		}
		return UserDTO{}, err
	}
	return NewUserDTO(user), nil
}

func (s *UserService) UpdateUser(id int64, input UpdateUserInput) (UserDTO, error) {
	existing, ok := s.store.GetUser(id)
	if !ok {
		return UserDTO{}, ErrNotFound
	}
	if input.Login != nil {
		login := strings.TrimSpace(*input.Login)
		if login == "" {
			return UserDTO{}, ValidationErrors{{
				Field:   "login",
				Message: "login cannot be empty",
			}}
		}
		if other, found := s.store.GetUserByLogin(login); found && other.ID != existing.ID {
			return UserDTO{}, ValidationErrors{{
				Field:   "login",
				Message: "user with this login already exists",
			}}
		}
		existing.Login = login
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return UserDTO{}, ValidationErrors{{
				Field:   "name",
				Message: "name cannot be empty",
			}}
		}
		existing.Name = name
	}
	if input.Role != nil {
		if !input.Role.Valid() {
			return UserDTO{}, ValidationErrors{{
				Field:   "role",
				Message: "invalid role",
			}}
		}
		existing.Role = *input.Role
	}
	if input.Password != nil {
		if len(*input.Password) < 6 {
			return UserDTO{}, ValidationErrors{{
				Field:   "password",
				Message: "password must be at least 6 characters",
			}}
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*input.Password), s.passwordCost)
		if err != nil {
			return UserDTO{}, err
		}
		existing.PasswordHash = string(hash)
	}
	updated, err := s.store.UpdateUser(*existing)
	if err != nil {
		if err == storage.ErrConflict {
			return UserDTO{}, ValidationErrors{{
				Field:   "login",
				Message: "user with this login already exists",
			}}
		}
		return UserDTO{}, err
	}
	return NewUserDTO(updated), nil
}

func (s *UserService) DeleteUser(id int64) error {
	if err := s.store.DeleteUser(id); err != nil {
		if err == storage.ErrNotFound {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func validateUserInput(input CreateUserInput) []ValidationError {
	var errs []ValidationError
	if strings.TrimSpace(input.Login) == "" {
		errs = append(errs, ValidationError{Field: "login", Message: "login is required"})
	}
	if len(input.Login) > 50 {
		errs = append(errs, ValidationError{Field: "login", Message: "login is too long"})
	}
	if strings.TrimSpace(input.Password) == "" {
		errs = append(errs, ValidationError{Field: "password", Message: "password is required"})
	}
	if len(input.Password) < 6 {
		errs = append(errs, ValidationError{Field: "password", Message: "password must be at least 6 characters"})
	}
	if strings.TrimSpace(input.Name) == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "name is required"})
	}
	if !input.Role.Valid() {
		errs = append(errs, ValidationError{Field: "role", Message: "invalid role"})
	}
	return errs
}
