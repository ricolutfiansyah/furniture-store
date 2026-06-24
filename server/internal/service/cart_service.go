package service

import (
	"context"
	"errors"
	"furniture-api/internal/domain"
)

type CartRepository interface {
	GetOrCreateCart(ctx context.Context, UserID int) (*domain.CartItem, error)
	GetCartWithItems(ctx context.Context, userID int) (*domain.Cart, error)
	AddItem(ctx context.Context, cartID, variantID, quantity int, priceAtTime float64) error
	UpdateItemQuantity(ctx context.Context, cartItemID, quantity int) error
	RemoveItem(ctx context.Context, cartItemID int) error
	ClearCart(ctx context.Context, cartID int) error
}

type ProductVariantRepository interface {
	GetByID(ctx context.Context, id int) (*domain.ProductVariant, error)
}

type CartService struct {
	cartRepo    CartRepository
	variantRepo ProductVariantRepository
}

func NewCartService(cartRepo CartRepository, variantRepo ProductVariantRepository) *CartService {
	return &CartService{cartRepo: cartRepo, variantRepo: variantRepo}
}

type AddToCartRequest struct {
	VariantID int `json:"variant_id"`
	Quantity  int `json:"quantity"`
}

func (s *CartService) AddToCart(ctx context.Context, userID int, req *AddToCartRequest) error {
	if req.Quantity <= 0 {
		return errors.New("Quantity must be greater than 0")
	}

	variant, err := s.variantRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("Variant not found")
	}
	if variant.StockQuantity < req.Quantity {
		return errors.New("Insufficient stock")
	}

	cart, err := s.cartRepo.GetOrCreateCart(ctx, userID)
	if err != nil {
		return err
	}

	totalPrice := variant.AdditionalPrice

	err = s.cartRepo.AddItem(ctx, cart.ID, req.VariantID, req.Quantity, totalPrice)
	return err
}
