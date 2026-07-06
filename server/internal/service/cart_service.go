package service

import (
	"context"
	"errors"
	"fmt"
	"furniture-api/internal/domain"
	"furniture-api/internal/repository"
)

type CartRepository interface {
	GetOrCreateCart(ctx context.Context, userID int) (*domain.Cart, error)
	GetCartWithItems(ctx context.Context, userID int) (*domain.Cart, error)
	AddItem(ctx context.Context, cartID, variantID, quantity int, priceAtTime float64) error
	UpdateItemQuantity(ctx context.Context, userID, cartItemID, quantity int) error
	RemoveItem(ctx context.Context, userID, cartItemID int) error
}

type ProductVariantRepository interface {
	GetVariantByID(ctx context.Context, id int) (*domain.ProductVariant, error)
}

type ProductRepositoryForCart interface {
	GetByID(ctx context.Context, id int) (*domain.Product, error)
}

type CartService struct {
	cartRepo    CartRepository
	variantRepo ProductVariantRepository
	productRepo ProductRepositoryForCart
}

func NewCartService(cartRepo CartRepository, variantRepo ProductVariantRepository, productRepo ProductRepositoryForCart) *CartService {
	return &CartService{cartRepo: cartRepo, variantRepo: variantRepo, productRepo: productRepo}
}

func (s *CartService) AddToCart(ctx context.Context, userID int, req *domain.AddToCartRequest) error {
	if req.Quantity <= 0 {
		return ErrInvalidQuantity
	}

	variant, err := s.variantRepo.GetVariantByID(ctx, req.VariantID)
	if err != nil {
		if errors.Is(err, repository.ErrVariantNotFound) {
			return ErrVariantNotFound
		}
		return fmt.Errorf("get variant: %w", err)
	}

	if variant.StockQuantity < req.Quantity {
		return ErrInsufficientStock
	}

	product, err := s.productRepo.GetByID(ctx, variant.ProductID)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return ErrVariantNotFound
		}
		return fmt.Errorf("get product: %w", err)
	}

	PriceAtTime := product.BasePrice + variant.AdditionalPrice

	cart, err := s.cartRepo.GetOrCreateCart(ctx, userID)
	if err != nil {
		return fmt.Errorf("get or create cart: %w", err)
	}

	if err = s.cartRepo.AddItem(ctx, cart.ID, req.VariantID, req.Quantity, PriceAtTime); err != nil {
		return fmt.Errorf("add item to cart: %w", err)
	}

	return nil
}

func (s *CartService) GetCart(ctx context.Context, userID int) (*domain.Cart, error) {
	cart, err := s.cartRepo.GetCartWithItems(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get cart: %w", err)
	}

	return cart, nil
}

func (s *CartService) UpdateQuantity(ctx context.Context, userID, cartItemID, quantity int) error {
	if quantity <= 0 {
		if err := s.cartRepo.RemoveItem(ctx, userID, cartItemID); err != nil {
			if errors.Is(err, repository.ErrCartItemNotFound) {
				return ErrCartItemNotFound
			}
			return fmt.Errorf("remove item: %w", err)
		}
		return nil
	}

	if err := s.cartRepo.UpdateItemQuantity(ctx, userID, cartItemID, quantity); err != nil {
		if errors.Is(err, repository.ErrCartItemNotFound) {
			return ErrCartItemNotFound
		}
		return fmt.Errorf("update item quantity: %w", err)
	}

	return nil
}

func (s *CartService) RemoveItem(ctx context.Context, userID, cartItemID int) error {
	if err := s.cartRepo.RemoveItem(ctx, userID, cartItemID); err != nil {
		if errors.Is(err, repository.ErrCartItemNotFound) {
			return ErrCartItemNotFound
		}
		return fmt.Errorf("remove item: %w", err)
	}

	return nil
}
