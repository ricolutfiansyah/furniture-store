package service

import (
	"context"
	"errors"
	"fmt"
	"furniture-api/internal/domain"
	"time"

	"github.com/jmoiron/sqlx"
)

type CartRepositoryForOrder interface {
	GetCartItemsByUserID(ctx context.Context, userID int) ([]domain.CartItem, error)
	ClearCart(ctx context.Context, cartID int) error
	GetCartByUserID(ctx context.Context, userID int) (*domain.Cart, error)
}

type ProductVariantRepositoryForOrder interface {
	GetByID(ctx context.Context, id int) (*domain.ProductVariant, error)
	DecreaseStockWithTx(ctx context.Context, tx *sqlx.Tx, variantID, quantity int) error
}

type OrderRepository interface {
	CreateOrderWithTx(ctx context.Context, tx *sqlx.Tx, order *domain.Order) error
	CreateOrderItemWithTx(ctx context.Context, tx *sqlx.Tx, item *domain.OrderItem) error
	CreateOrderStatusWithTx(ctx context.Context, tx *sqlx.Tx, orderID int, status, notes, createdBy string) error
	GetOrdersByUserID(ctx context.Context, userID int) ([]domain.Order, error)
	GetOrderByID(ctx context.Context, orderID int) (*domain.Order, error)
	GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]domain.OrderItem, error)
	GetOrderStatusesByOrderID(ctx context.Context, orderID int) ([]domain.OrderStatus, error)
}

type OrderService struct {
	orderRepo   OrderRepository
	cartRepo    CartRepositoryForOrder
	variantRepo ProductVariantRepositoryForOrder
	db          *sqlx.DB
}

func NewOrderService(
	orderRepo OrderRepository,
	cartRepo CartRepositoryForOrder,
	variantRepo ProductVariantRepositoryForOrder,
	db *sqlx.DB,
) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		variantRepo: variantRepo,
		db:          db,
	}
}

type CheckoutRequest struct {
	ShippingAddress string `json:"shipping_address"`
	Notes           string `json:"notes"`
}

type CheckoutResponse struct {
	Order      domain.Order       `json:"order"`
	Items      []domain.OrderItem `json:"items"`
	GrandTotal float64            `json:"grand_total"`
}

func (s *OrderService) Checkout(ctx context.Context, userID int, req *CheckoutRequest) (*CheckoutResponse, error) {
	cartItems, err := s.cartRepo.GetCartItemsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(cartItems) == 0 {
		return nil, errors.New("Cart is empty")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	orderNumber := fmt.Sprintf("ORD-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)

	var totalAmount, grandTotal float64
	var orderItems []domain.OrderItem

	for _, cartItem := range cartItems {
		variant, err := s.variantRepo.GetByID(ctx, cartItem.VariantID)
		if err != nil || variant == nil {
			return nil, fmt.Errorf("Variant %d not found", cartItem.VariantID)
		}

		if variant.StockQuantity < cartItem.Quantity {
			return nil, fmt.Errorf("Insufficient stock for variant %s (Available: %d, requested: %d)", variant.VariantName, variant.StockQuantity, cartItem.Quantity)
		}

		pricePerItem := cartItem.PriceAtTime

		totalItemPrice := float64(cartItem.Quantity) * pricePerItem
		totalAmount += totalItemPrice

		orderItem := domain.OrderItem{
			VariantID:    cartItem.VariantID,
			Quantity:     cartItem.Quantity,
			PricePerItem: pricePerItem,
			TotalPrice:   totalItemPrice,
		}
		orderItems = append(orderItems, orderItem)

		err = s.variantRepo.DecreaseStockWithTx(ctx, tx, cartItem.VariantID, cartItem.Quantity)
		if err != nil {
			return nil, err
		}
	}

	shippingCost := 0.0
	tax := totalAmount * 0.12
	grandTotal = totalAmount + shippingCost + tax

	order := &domain.Order{
		UserID:          userID,
		OrderNumber:     orderNumber,
		TotalAmount:     totalAmount,
		ShippingCost:    shippingCost,
		Tax:             tax,
		GrandTotal:      grandTotal,
		Status:          "pending",
		ShippingAddress: req.ShippingAddress,
		PaymentMethod:   "bank_transfer",
		Notes:           req.Notes,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err = s.orderRepo.CreateOrderWithTx(ctx, tx, order)
	if err != nil {
		return nil, err
	}

	for i := range orderItems {
		orderItems[i].OrderID = order.ID
		err = s.orderRepo.CreateOrderItemWithTx(ctx, tx, &orderItems[i])
		if err != nil {
			return nil, err
		}
	}

	err = s.orderRepo.CreateOrderStatusWithTx(ctx, tx, order.ID, "pending", "Order created", "System")
	if err != nil {
		return nil, err
	}

	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if cart != nil {
		err = s.cartRepo.ClearCart(ctx, cart.ID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &CheckoutResponse{
		Order:      *order,
		Items:      orderItems,
		GrandTotal: grandTotal,
	}, nil
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID int) ([]domain.Order, error) {
	return s.orderRepo.GetOrdersByUserID(ctx, userID)
}

func (s *OrderService) GetOrderDetail(ctx context.Context, userID, orderID int) (*domain.Order, error) {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("Order not found")
	}

	if order.UserID != userID {
		return nil, errors.New("Unauthorized")
	}

	items, err := s.orderRepo.GetOrderItemsByOrderID(ctx, orderID)
	if err == nil {
		order.Items = items
	}

	statuses, err := s.orderRepo.GetOrderStatusesByOrderID(ctx, orderID)
	if err == nil {
		order.Statuses = statuses
	}

	return order, nil
}
