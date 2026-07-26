package service_test

import (
	"context"
	"errors"
	"furniture-api/internal/domain"
	"furniture-api/internal/nullable"
	"furniture-api/internal/service"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// --- mock ---

type mockUserRepo struct {
	findByEmailFn    func(ctx context.Context, email string) (*domain.User, error)
	createFn         func(ctx context.Context, user *domain.User) error
	findByIdFn       func(ctx context.Context, id int) (*domain.User, error)
	findByPublicIDFn func(ctx context.Context, publicID string) (*domain.User, error)
	updateFn         func(ctx context.Context, user *domain.User) error
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.findByEmailFn(ctx, email)
}
func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	return m.createFn(ctx, user)
}
func (m *mockUserRepo) FindById(ctx context.Context, id int) (*domain.User, error) {
	return m.findByIdFn(ctx, id)
}
func (m *mockUserRepo) FindByPublicID(ctx context.Context, publicID string) (*domain.User, error) {
	return m.findByPublicIDFn(ctx, publicID)
}
func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	return m.updateFn(ctx, user)
}

// --- helpers ---

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		findByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
			return nil, errors.New("FindByEmail: unexpected call")
		},
		createFn: func(ctx context.Context, user *domain.User) error {
			return errors.New("Create: unexpected call")
		},
		findByIdFn: func(ctx context.Context, id int) (*domain.User, error) {
			return nil, errors.New("FindById: unexpected call")
		},
		findByPublicIDFn: func(ctx context.Context, publicID string) (*domain.User, error) {
			return nil, errors.New("FindByPublicID: unexpected call")
		},
		updateFn: func(ctx context.Context, user *domain.User) error {
			return errors.New("Update: unexpected call")
		},
	}
}

func fixUser(email string) *domain.User {
	return &domain.User{
		ID:           1,
		PublicID:     uuid.New().String(),
		Email:        email,
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqR9U2v.9Q1M4x9jXjxTV0YQ4LgLW",
		FullName:     nullable.NewNullString("Test User"),
		Role:         "user",
		IsActive:     true,
	}
}

// --- Register ---

func TestRegister_Success(t *testing.T) {
	mock := newMockUserRepo()
	mock.findByEmailFn = func(ctx context.Context, email string) (*domain.User, error) {
		return nil, domain.ErrUserNotFound
	}
	mock.createFn = func(ctx context.Context, user *domain.User) error {
		return nil
	}

	svc := service.NewAuthService(mock, "test-secret")
	user, err := svc.Register(context.Background(), &domain.RegisterRequest{
		Email:    "new@example.com",
		Password: "password123",
		FullName: "New User",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != "new@example.com" {
		t.Errorf("expected email new@example.com, got %s", user.Email)
	}
	if user.Role != "user" {
		t.Errorf("expected role 'user', got %s", user.Role)
	}
	if user.PublicID == "" {
		t.Error("expected public_id to be set")
	}
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	mock := newMockUserRepo()
	mock.findByEmailFn = func(ctx context.Context, email string) (*domain.User, error) {
		return fixUser(email), nil
	}

	svc := service.NewAuthService(mock, "test-secret")
	_, err := svc.Register(context.Background(), &domain.RegisterRequest{
		Email:    "exists@example.com",
		Password: "password123",
	})

	if !errors.Is(err, domain.ErrEmailAlreadyRegistered) {
		t.Fatalf("expected ErrEmailAlreadyRegistered, got %v", err)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	svc := service.NewAuthService(newMockUserRepo(), "test-secret")
	_, err := svc.Register(context.Background(), &domain.RegisterRequest{
		Email:    "not-an-email",
		Password: "password123",
	})

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	svc := service.NewAuthService(newMockUserRepo(), "test-secret")
	_, err := svc.Register(context.Background(), &domain.RegisterRequest{
		Email:    "test@example.com",
		Password: "123",
	})

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestRegister_RepoFindByEmailError(t *testing.T) {
	mock := newMockUserRepo()
	mock.findByEmailFn = func(ctx context.Context, email string) (*domain.User, error) {
		return nil, errors.New("db connection lost")
	}

	svc := service.NewAuthService(mock, "test-secret")
	_, err := svc.Register(context.Background(), &domain.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRegister_RepoCreateError(t *testing.T) {
	mock := newMockUserRepo()
	mock.findByEmailFn = func(ctx context.Context, email string) (*domain.User, error) {
		return nil, domain.ErrUserNotFound
	}
	mock.createFn = func(ctx context.Context, user *domain.User) error {
		return domain.ErrEmailAlreadyRegistered // race condition: duplicate between check and insert
	}

	svc := service.NewAuthService(mock, "test-secret")
	_, err := svc.Register(context.Background(), &domain.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	if !errors.Is(err, domain.ErrEmailAlreadyRegistered) {
		t.Fatalf("expected ErrEmailAlreadyRegistered, got %v", err)
	}
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := fixUser("test@example.com")
	user.PasswordHash = string(hashed)

	mock := newMockUserRepo()
	mock.findByEmailFn = func(ctx context.Context, email string) (*domain.User, error) {
		return user, nil
	}

	svc := service.NewAuthService(mock, "test-secret")
	resp, err := svc.Login(context.Background(), &domain.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Token == "" {
		t.Error("expected token to be set")
	}
	if resp.User.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", resp.User.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := fixUser("test@example.com")
	user.PasswordHash = string(hashed)

	mock := newMockUserRepo()
	mock.findByEmailFn = func(ctx context.Context, email string) (*domain.User, error) {
		return user, nil
	}

	svc := service.NewAuthService(mock, "test-secret")
	_, err := svc.Login(context.Background(), &domain.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})

	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	mock := newMockUserRepo()
	mock.findByEmailFn = func(ctx context.Context, email string) (*domain.User, error) {
		return nil, domain.ErrUserNotFound
	}

	svc := service.NewAuthService(mock, "test-secret")
	_, err := svc.Login(context.Background(), &domain.LoginRequest{
		Email:    "ghost@example.com",
		Password: "password123",
	})

	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_InactiveUser(t *testing.T) {
	mock := newMockUserRepo()
	mock.findByEmailFn = func(ctx context.Context, email string) (*domain.User, error) {
		u := fixUser(email)
		u.IsActive = false
		return u, nil
	}

	svc := service.NewAuthService(mock, "test-secret")
	_, err := svc.Login(context.Background(), &domain.LoginRequest{
		Email:    "banned@example.com",
		Password: "password123",
	})

	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_InvalidEmail(t *testing.T) {
	svc := service.NewAuthService(newMockUserRepo(), "test-secret")
	_, err := svc.Login(context.Background(), &domain.LoginRequest{
		Email:    "bad",
		Password: "password123",
	})

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

// --- GetProfile ---

func TestGetProfile_Success(t *testing.T) {
	user := fixUser("test@example.com")

	mock := newMockUserRepo()
	mock.findByPublicIDFn = func(ctx context.Context, publicID string) (*domain.User, error) {
		return user, nil
	}

	svc := service.NewAuthService(mock, "test-secret")
	resp, err := svc.GetProfile(context.Background(), user.PublicID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, resp.Email)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	mock := newMockUserRepo()
	mock.findByPublicIDFn = func(ctx context.Context, publicID string) (*domain.User, error) {
		return nil, domain.ErrUserNotFound
	}

	svc := service.NewAuthService(mock, "test-secret")
	_, err := svc.GetProfile(context.Background(), "non-existent")

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
