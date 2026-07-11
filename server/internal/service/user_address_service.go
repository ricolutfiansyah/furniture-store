package service

import (
	"context"
	"fmt"
	"furniture-api/internal/domain"
	"furniture-api/internal/nullable"
	"furniture-api/internal/validation"
	"strings"

	"github.com/jmoiron/sqlx"
)

type AddressRepository interface {
	Create(ctx context.Context, address *domain.UserAddress) error
	CountByUserID(ctx context.Context, userID int) (int, error)
	GetByID(ctx context.Context, id, userID int) (*domain.UserAddress, error)
	ListByUserID(ctx context.Context, userID int) ([]domain.UserAddress, error)
	Update(ctx context.Context, address *domain.UserAddress) error
	Delete(ctx context.Context, id, userID int) error
	UnsetDefaultByUserID(ctx context.Context, tx *sqlx.Tx, userID int) error
	SetDefault(ctx context.Context, tx *sqlx.Tx, id, userID int) error
	ListRemainingAfterDelete(ctx context.Context, userID, excludeID int) ([]domain.UserAddress, error)
}

type addressService struct {
	addressRepo AddressRepository
	db          *sqlx.DB
}

func NewAddressService(r AddressRepository, db *sqlx.DB) *addressService {
	return &addressService{addressRepo: r, db: db}
}

func (s *addressService) CreateAddress(ctx context.Context, userID int, req domain.CreateAddressRequest) (*domain.UserAddress, error) {
	req.RecipientName = strings.TrimSpace(req.RecipientName)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Province = strings.TrimSpace(req.Province)
	req.City = strings.TrimSpace(req.City)
	req.District = strings.TrimSpace(req.District)
	req.PostalCode = strings.TrimSpace(req.PostalCode)
	req.AddressLine = strings.TrimSpace(req.AddressLine)

	if err := validation.Validate(
		validation.Required("recipient_name", req.RecipientName),
		validation.Required("phone", req.Phone),
		validation.Required("province", req.Province),
		validation.Required("city", req.City),
		validation.Required("district", req.District),
		validation.Required("postal_code", req.PostalCode),
		validation.Required("address_line", req.AddressLine),
	); err != nil {
		return nil, err
	}

	existingCount, err := s.addressRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count existing addresses: %w", err)
	}

	var label nullable.NullString
	if req.Label != nil {
		label = nullable.NewNullString(strings.TrimSpace(*req.Label))
	}

	address := &domain.UserAddress{
		UserID:        userID,
		Label:         label,
		RecipientName: req.RecipientName,
		Phone:         req.Phone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		PostalCode:    req.PostalCode,
		AddressLine:   req.AddressLine,
		IsDefault:     existingCount == 0,
	}

	if err := s.addressRepo.Create(ctx, address); err != nil {
		return nil, fmt.Errorf("create address: %w", err)
	}

	return address, nil
}
