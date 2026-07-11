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
	ListByUserIDTx(ctx context.Context, tx *sqlx.Tx, userID int) ([]domain.UserAddress, error)
	Update(ctx context.Context, address *domain.UserAddress) error

	GetByIDTx(ctx context.Context, tx *sqlx.Tx, id, userID int) (*domain.UserAddress, error)
	DeleteTx(ctx context.Context, tx *sqlx.Tx, id, userID int) error
	UnsetDefaultByUserID(ctx context.Context, tx *sqlx.Tx, userID int) error
	SetDefault(ctx context.Context, tx *sqlx.Tx, id, userID int) error
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

func (s *addressService) UpdateAddress(ctx context.Context, id, userID int, req domain.UpdateAddressRequest) (*domain.UserAddress, error) {
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

	existing, err := s.addressRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	var label nullable.NullString
	if req.Label != nil {
		label = nullable.NewNullString(strings.TrimSpace(*req.Label))
	}

	existing.Label = label
	existing.RecipientName = req.RecipientName
	existing.Phone = req.Phone
	existing.Province = req.Province
	existing.City = req.City
	existing.District = req.District
	existing.PostalCode = req.PostalCode
	existing.AddressLine = req.AddressLine

	if err := s.addressRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("update address: %w", err)
	}

	return existing, nil
}

func (s *addressService) DeleteAddress(ctx context.Context, id, userID int) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	existing, err := s.addressRepo.GetByIDTx(ctx, tx, id, userID)
	if err != nil {
		return err
	}

	if err := s.addressRepo.DeleteTx(ctx, tx, id, userID); err != nil {
		return fmt.Errorf("delete address: %w", err)
	}

	if !existing.IsDefault {
		return tx.Commit()
	}

	remaining, err := s.addressRepo.ListByUserIDTx(ctx, tx, userID)
	if err != nil {
		return fmt.Errorf("list remaining address: %w", err)
	}

	switch len(remaining) {
	case 0:
		// no remaining address, nothing to set to default
	case 1:
		if err := s.addressRepo.SetDefault(ctx, tx, remaining[0].ID, userID); err != nil {
			return fmt.Errorf("promote remaining address to default: %w", err)
		}
	default:
		// >1 remaining, no default auto-set
	}

	return tx.Commit()
}

func (s *addressService) ListAddresses(ctx context.Context, userID int) ([]domain.UserAddress, error) {
	addresses, err := s.addressRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list addresses: %w", err)
	}

	return addresses, nil
}

func (s *addressService) SetDefaultAddress(ctx context.Context, id, userID int) error {
	existing, err := s.addressRepo.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}
	if existing.IsDefault {
		return nil
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := s.addressRepo.GetByIDTx(ctx, tx, id, userID); err != nil {
		return nil
	}

	if err := s.addressRepo.UnsetDefaultByUserID(ctx, tx, userID); err != nil {
		return fmt.Errorf("unser current default: %w", err)
	}

	if err := s.addressRepo.SetDefault(ctx, tx, id, userID); err != nil {
		return fmt.Errorf("set new default: %w", err)
	}

	return tx.Commit()
}
