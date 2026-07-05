package service

import (
	"context"
	"errors"
	"fmt"
	"furniture-api/internal/domain"
	"furniture-api/internal/repository"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	FindById(ctx context.Context, id int) (*domain.User, error)
	FindByPublicID(ctx context.Context, publicID string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

type AuthService struct {
	userRepo  UserRepository
	jwtSecret string
}

func NewAuthService(userRepo UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) GetProfile(ctx context.Context, publicID string) (*UserResponse, error) {
	user, err := s.userRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, err
	}

	resp := toUserResponse(user)
	return &resp, nil
}

func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*domain.User, error) {
	req.Password = strings.ToLower(strings.TrimSpace(req.Password))
	req.FullName = strings.TrimSpace(req.FullName)

	if len(req.Password) < 8 {
		return nil, ErrPasswordTooShort
	}

	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("check existing email: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyRegistered
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		PublicID:     uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hashed),
		FullName:     toNullString(req.FullName),
		Phone:        toNullString(req.Phone),
		Address:      toNullString(req.Address),
		Role:         "user",
		IsActive:     true,
	}

	if err = s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyRegistered) {
			return nil, ErrEmailAlreadyRegistered
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// consistency response time
const dummyHash = "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqR9U2v.9Q1M4x9jXjxTV0YQ4LgLW"

func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(req.Password))
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("fund user by email: %w", err)
	}

	if !user.IsActive {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.PublicID,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"email": user.Email,
		"role":  user.Role,
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}

	return &LoginResponse{
		Token: tokenString,
		User:  toUserResponse(user),
	}, nil
}
