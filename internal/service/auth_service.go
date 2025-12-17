package service

import (
	"appointment-booking/internal/domain"
	"appointment-booking/internal/repository"
	"appointment-booking/pkg/utils"
	"errors"
)

type AuthService struct {
	repo *repository.UserRepository
}

func NewAuthService(repo *repository.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"` // Optional, default to customer
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (s *AuthService) Register(input RegisterInput) error {
	// 1. Hash Password
	hashedPwd, err := utils.HashPassword(input.Password)
	if err != nil {
		return err
	}

	// 2. Set Default Role if empty
	role := domain.RoleCustomer
	if input.Role == "provider" {
		role = domain.RoleProvider
	}
	// Security: Only admins should be able to create "admin" users via API.
	// For now, we allow it for simplicity, but note this for interviews.

	// 3. Create User
	user := domain.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: hashedPwd,
		Role:     role,
	}

	return s.repo.Create(&user)
}

func (s *AuthService) Login(input LoginInput) (string, error) {
	// 1. Find User
	user, err := s.repo.FindByEmail(input.Email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// 2. Check Password
	if !utils.CheckPassword(input.Password, user.Password) {
		return "", errors.New("invalid credentials")
	}

	// 3. Generate Token
	token, err := utils.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		return "", err
	}

	return token, nil
}
