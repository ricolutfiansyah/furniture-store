package service

import (
	"context"
	"errors"
	"fmt"
	"furniture-api/internal/domain"
	"furniture-api/internal/nullable"
	"furniture-api/internal/repository"
	"furniture-api/internal/validation"
	"strings"
	"time"

	"regexp"

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

func NewAuthService(r UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  r,
		jwtSecret: jwtSecret,
	}
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (s *AuthService) GetProfile(ctx context.Context, publicID string) (*domain.UserResponse, error) {
	user, err := s.userRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, err
	}

	resp := domain.ToUserResponse(user)
	return &resp, nil
}

func (s *AuthService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.FullName = strings.TrimSpace(req.FullName)

	if err := validation.Validate(
		validation.Required("email", req.Email),
		validation.IsValidEmail("email", req.Email),
		validation.Required("password", req.Password),
		validation.MinLength("password", req.Password, 8),
	); err != nil {
		return nil, err
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
		FullName:     nullable.ToNullString(req.FullName),
		Phone:        nullable.ToNullString(req.Phone),
		Address:      nullable.ToNullString(req.Address),
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

func (s *AuthService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if err := validation.Validate(
		validation.Required("email", req.Email),
		validation.IsValidEmail("email", req.Email),
		validation.Required("password", req.Password),
		validation.MinLength("password", req.Password, 8),
	); err != nil {
		return nil, err
	}

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

	return &domain.LoginResponse{
		Token: tokenString,
		User:  domain.ToUserResponse(user),
	}, nil
}
