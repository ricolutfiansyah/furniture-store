package service_test

import (
	"context"
	"errors"
	"furniture-api/internal/domain"
	"furniture-api/internal/service"
	"testing"
)

// --- cart mock ---

type mockCartRepo struct {
	getOrCreateCartFn    func(ctx context.Context, userID int) (*domain.Cart, error)
	getCartWithItemsFn   func(ctx context.Context, userID int) (*domain.Cart, error)
	addItemFn            func(ctx context.Context, cartID, variantID, quantity int, priceAtTime float64) error
	updateItemQuantityFn func(ctx context.Context, userID, cartItemID, quantity int) error
	removeItemFn         func(ctx context.Context, userID, cartItemID int) error
	removeItemsFn        func(ctx context.Context, userID int, itemIDs []int) error
	getCartItemFn        func(ctx context.Context, cartID, variantID int) (*domain.CartItem, error)
	getCartItemByIDFn    func(ctx context.Context, userID, cartItemID int) (*domain.CartItem, error)
}

func (m *mockCartRepo) GetOrCreateCart(ctx context.Context, userID int) (*domain.Cart, error) {
	return m.getOrCreateCartFn(ctx, userID)
}
func (m *mockCartRepo) GetCartWithItems(ctx context.Context, userID int) (*domain.Cart, error) {
	return m.getCartWithItemsFn(ctx, userID)
}
func (m *mockCartRepo) AddItem(ctx context.Context, cartID, variantID, quantity int, priceAtTime float64) error {
	return m.addItemFn(ctx, cartID, variantID, quantity, priceAtTime)
}
func (m *mockCartRepo) UpdateItemQuantity(ctx context.Context, userID, cartItemID, quantity int) error {
	return m.updateItemQuantityFn(ctx, userID, cartItemID, quantity)
}
func (m *mockCartRepo) RemoveItem(ctx context.Context, userID, cartItemID int) error {
	return m.removeItemFn(ctx, userID, cartItemID)
}
func (m *mockCartRepo) RemoveItems(ctx context.Context, userID int, itemIDs []int) error {
	return m.removeItemsFn(ctx, userID, itemIDs)
}
func (m *mockCartRepo) GetCartItem(ctx context.Context, cartID, variantID int) (*domain.CartItem, error) {
	return m.getCartItemFn(ctx, cartID, variantID)
}
func (m *mockCartRepo) GetCartItemByID(ctx context.Context, userID, cartItemID int) (*domain.CartItem, error) {
	return m.getCartItemByIDFn(ctx, userID, cartItemID)
}

// --- product catalog mock ---

type mockProductCatalogRepo struct {
	getVariantByIDFn func(ctx context.Context, id int) (*domain.ProductVariant, error)
	getByIDFn        func(ctx context.Context, id int) (*domain.Product, error)
}

func (m *mockProductCatalogRepo) GetVariantByID(ctx context.Context, id int) (*domain.ProductVariant, error) {
	return m.getVariantByIDFn(ctx, id)
}
func (m *mockProductCatalogRepo) GetByID(ctx context.Context, id int) (*domain.Product, error) {
	return m.getByIDFn(ctx, id)
}

// --- helpers ---

func defaultMockCartRepo() *mockCartRepo {
	return &mockCartRepo{
		getOrCreateCartFn: func(ctx context.Context, userID int) (*domain.Cart, error) {
			return &domain.Cart{ID: 10, UserID: userID}, nil
		},
		addItemFn: func(ctx context.Context, cartID, variantID, quantity int, priceAtTime float64) error {
			return nil
		},
		getCartItemFn: func(ctx context.Context, cartID, variantID int) (*domain.CartItem, error) {
			return nil, domain.ErrCartItemNotFound
		},
	}
}

func defaultMockProductCatalogRepo() *mockProductCatalogRepo {
	return &mockProductCatalogRepo{
		getVariantByIDFn: func(ctx context.Context, id int) (*domain.ProductVariant, error) {
			return &domain.ProductVariant{
				ID:              id,
				ProductID:       1,
				VariantName:    "Test Variant",
				AdditionalPrice: 5000,
				StockQuantity:   100,
				IsActive:        true,
			}, nil
		},
		getByIDFn: func(ctx context.Context, id int) (*domain.Product, error) {
			return &domain.Product{
				ID:        id,
				Name:     "Test Product",
				BasePrice: 50000,
				IsActive: true,
			}, nil
		},
	}
}

func newCartService(cartRepo *mockCartRepo, catalogRepo *mockProductCatalogRepo) *service.CartService {
	return service.NewCartService(cartRepo, catalogRepo)
}

// --- AddToCart ---

func TestAddToCart_Success(t *testing.T) {
	cartRepo := defaultMockCartRepo()
	catalogRepo := defaultMockProductCatalogRepo()

	svc := newCartService(cartRepo, catalogRepo)
	err := svc.AddToCart(context.Background(), 1, &domain.AddToCartRequest{
		VariantID: 5,
		Quantity:  2,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestAddToCart_InvalidQuantity(t *testing.T) {
	svc := newCartService(defaultMockCartRepo(), defaultMockProductCatalogRepo())
	err := svc.AddToCart(context.Background(), 1, &domain.AddToCartRequest{
		VariantID: 5,
		Quantity:  0,
	})

	if !errors.Is(err, domain.ErrInvalidQuantity) {
		t.Fatalf("expected ErrInvalidQuantity, got %v", err)
	}
}

func TestAddToCart_InvalidVariantID(t *testing.T) {
	svc := newCartService(defaultMockCartRepo(), defaultMockProductCatalogRepo())
	err := svc.AddToCart(context.Background(), 1, &domain.AddToCartRequest{
		VariantID: 0,
		Quantity:  2,
	})

	if !errors.Is(err, domain.ErrInvalidVariantID) {
		t.Fatalf("expected ErrInvalidVariantID, got %v", err)
	}
}

func TestAddToCart_VariantNotFound(t *testing.T) {
	catalogRepo := defaultMockProductCatalogRepo()
	catalogRepo.getVariantByIDFn = func(ctx context.Context, id int) (*domain.ProductVariant, error) {
		return nil, domain.ErrVariantNotFound
	}

	svc := newCartService(defaultMockCartRepo(), catalogRepo)
	err := svc.AddToCart(context.Background(), 1, &domain.AddToCartRequest{
		VariantID: 999,
		Quantity:  1,
	})

	if !errors.Is(err, domain.ErrVariantNotFound) {
		t.Fatalf("expected ErrVariantNotFound, got %v", err)
	}
}

func TestAddToCart_InsufficientStock(t *testing.T) {
	catalogRepo := defaultMockProductCatalogRepo()
	catalogRepo.getVariantByIDFn = func(ctx context.Context, id int) (*domain.ProductVariant, error) {
		return &domain.ProductVariant{
			ID:            5,
			StockQuantity: 0,
		}, nil
	}

	svc := newCartService(defaultMockCartRepo(), catalogRepo)
	err := svc.AddToCart(context.Background(), 1, &domain.AddToCartRequest{
		VariantID: 5,
		Quantity:  1,
	})

	if !errors.Is(err, domain.ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestAddToCart_ProductNotFound(t *testing.T) {
	catalogRepo := defaultMockProductCatalogRepo()
	catalogRepo.getByIDFn = func(ctx context.Context, id int) (*domain.Product, error) {
		return nil, domain.ErrProductNotFound
	}

	svc := newCartService(defaultMockCartRepo(), catalogRepo)
	err := svc.AddToCart(context.Background(), 1, &domain.AddToCartRequest{
		VariantID: 5,
		Quantity:  1,
	})

	if !errors.Is(err, domain.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestAddToCart_ExceedsStockWithExistingItem(t *testing.T) {
	catalogRepo := defaultMockProductCatalogRepo()
	catalogRepo.getVariantByIDFn = func(ctx context.Context, id int) (*domain.ProductVariant, error) {
		return &domain.ProductVariant{ID: 5, ProductID: 1, StockQuantity: 5}, nil
	}

	cartRepo := defaultMockCartRepo()
	cartRepo.getCartItemFn = func(ctx context.Context, cartID, variantID int) (*domain.CartItem, error) {
		return &domain.CartItem{ID: 1, Quantity: 4, VariantID: 5}, nil // already have 4
	}

	svc := newCartService(cartRepo, catalogRepo)
	err := svc.AddToCart(context.Background(), 1, &domain.AddToCartRequest{
		VariantID: 5,
		Quantity:  2, // 4 + 2 = 6 > 5 stock
	})

	if !errors.Is(err, domain.ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
}

// --- GetCart ---

func TestGetCart_Success(t *testing.T) {
	cartRepo := defaultMockCartRepo()
	cartRepo.getCartWithItemsFn = func(ctx context.Context, userID int) (*domain.Cart, error) {
		return &domain.Cart{
			ID:     10,
			UserID: userID,
			Items:  []domain.CartItem{{ID: 1, Quantity: 2}},
		}, nil
	}

	svc := newCartService(cartRepo, defaultMockProductCatalogRepo())
	cart, err := svc.GetCart(context.Background(), 1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cart.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(cart.Items))
	}
}

// --- UpdateQuantity ---

func TestUpdateQuantity_Success(t *testing.T) {
	catalogRepo := defaultMockProductCatalogRepo()
	catalogRepo.getVariantByIDFn = func(ctx context.Context, id int) (*domain.ProductVariant, error) {
		return &domain.ProductVariant{ID: 5, StockQuantity: 20}, nil
	}

	cartRepo := defaultMockCartRepo()
	cartRepo.getCartItemByIDFn = func(ctx context.Context, userID, cartItemID int) (*domain.CartItem, error) {
		return &domain.CartItem{ID: cartItemID, VariantID: 5, Quantity: 3}, nil
	}
	cartRepo.updateItemQuantityFn = func(ctx context.Context, userID, cartItemID, quantity int) error {
		return nil
	}

	svc := newCartService(cartRepo, catalogRepo)
	err := svc.UpdateQuantity(context.Background(), 1, 100, 5)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdateQuantity_RemoveOnZero(t *testing.T) {
	cartRepo := defaultMockCartRepo()
	cartRepo.removeItemFn = func(ctx context.Context, userID, cartItemID int) error {
		return nil
	}

	svc := newCartService(cartRepo, defaultMockProductCatalogRepo())
	err := svc.UpdateQuantity(context.Background(), 1, 100, 0)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdateQuantity_InsufficientStock(t *testing.T) {
	catalogRepo := defaultMockProductCatalogRepo()
	catalogRepo.getVariantByIDFn = func(ctx context.Context, id int) (*domain.ProductVariant, error) {
		return &domain.ProductVariant{ID: 5, StockQuantity: 2}, nil
	}

	cartRepo := defaultMockCartRepo()
	cartRepo.getCartItemByIDFn = func(ctx context.Context, userID, cartItemID int) (*domain.CartItem, error) {
		return &domain.CartItem{ID: cartItemID, VariantID: 5}, nil
	}

	svc := newCartService(cartRepo, catalogRepo)
	err := svc.UpdateQuantity(context.Background(), 1, 100, 10)

	if !errors.Is(err, domain.ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestUpdateQuantity_CartItemNotFound(t *testing.T) {
	cartRepo := defaultMockCartRepo()
	cartRepo.getCartItemByIDFn = func(ctx context.Context, userID, cartItemID int) (*domain.CartItem, error) {
		return nil, domain.ErrCartItemNotFound
	}

	svc := newCartService(cartRepo, defaultMockProductCatalogRepo())
	err := svc.UpdateQuantity(context.Background(), 1, 999, 5)

	if !errors.Is(err, domain.ErrCartItemNotFound) {
		t.Fatalf("expected ErrCartItemNotFound, got %v", err)
	}
}

// --- RemoveItem ---

func TestRemoveItem_Success(t *testing.T) {
	cartRepo := defaultMockCartRepo()
	cartRepo.removeItemFn = func(ctx context.Context, userID, cartItemID int) error {
		return nil
	}

	svc := newCartService(cartRepo, defaultMockProductCatalogRepo())
	err := svc.RemoveItem(context.Background(), 1, 100)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRemoveItem_NotFound(t *testing.T) {
	cartRepo := defaultMockCartRepo()
	cartRepo.removeItemFn = func(ctx context.Context, userID, cartItemID int) error {
		return domain.ErrCartItemNotFound
	}

	svc := newCartService(cartRepo, defaultMockProductCatalogRepo())
	err := svc.RemoveItem(context.Background(), 1, 999)

	if !errors.Is(err, domain.ErrCartItemNotFound) {
		t.Fatalf("expected ErrCartItemNotFound, got %v", err)
	}
}

// --- BulkRemoveItems ---

func TestBulkRemoveItems_Success(t *testing.T) {
	cartRepo := defaultMockCartRepo()
	cartRepo.removeItemsFn = func(ctx context.Context, userID int, itemIDs []int) error {
		return nil
	}

	svc := newCartService(cartRepo, defaultMockProductCatalogRepo())
	err := svc.BulkRemoveItems(context.Background(), 1, []int{1, 2, 3})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBulkRemoveItems_EmptyList(t *testing.T) {
	svc := newCartService(defaultMockCartRepo(), defaultMockProductCatalogRepo())
	err := svc.BulkRemoveItems(context.Background(), 1, []int{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
