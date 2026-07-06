package service

import (
	"context"
	"errors"
	"fmt"
	"furniture-api/internal/domain"
	"furniture-api/internal/nullable"
	"furniture-api/internal/repository"
	"log"
	"slices"
	"time"

	"github.com/jmoiron/sqlx"
)

type OrderRepository interface {
	CreateOrderWithTx(ctx context.Context, tx *sqlx.Tx, order *domain.Order) error
	CreateOrderItemWithTx(ctx context.Context, tx *sqlx.Tx, item *domain.OrderItem) error
	CreateOrderStatusWithTx(ctx context.Context, tx *sqlx.Tx, orderID int, status string, notes nullable.NullString, createdBy string) error
	UpdateOrderStatusWithTx(ctx context.Context, tx *sqlx.Tx, orderID int, status string, timestampColumn string) error

	GetOrdersByUserID(ctx context.Context, userID int) ([]domain.Order, error)
	GetOrderByID(ctx context.Context, userID, orderID int) (*domain.Order, error)
	GetOrderByIDForAdmin(ctx context.Context, orderID int) (*domain.Order, error)
	GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]domain.OrderItem, error)
	GetOrderStatusesByOrderID(ctx context.Context, orderID int) ([]domain.OrderStatus, error)
	GetOrderStatusForUpdate(ctx context.Context, tx *sqlx.Tx, orderID int) (string, error)
}

type CartRepositoryForOrder interface {
	GetCartItemsByUserIDTx(ctx context.Context, tx *sqlx.Tx, userID int) ([]domain.CartItem, error)
	ClearCartWithTx(ctx context.Context, tx *sqlx.Tx, cartID int) error
}

type ProductVariantRepositoryForOrder interface {
	GetVariantByID(ctx context.Context, id int) (*domain.ProductVariant, error)
	DecreaseStockWithTx(ctx context.Context, tx *sqlx.Tx, variantID, quantity int) error
}

type UserRepositoryForOrder interface {
	FindById(ctx context.Context, id int) (*domain.User, error)
}

type OrderService struct {
	orderRepo   OrderRepository
	cartRepo    CartRepositoryForOrder
	variantRepo ProductVariantRepositoryForOrder
	userRepo    UserRepositoryForOrder
	db          *sqlx.DB
}

func NewOrderService(
	orderRepo OrderRepository,
	cartRepo CartRepositoryForOrder,
	variantRepo ProductVariantRepositoryForOrder,
	userRepo UserRepositoryForOrder,
	db *sqlx.DB,
) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		variantRepo: variantRepo,
		userRepo:    userRepo,
		db:          db,
	}
}

const taxRate = 0.12
const maxOrderNumberAttempts = 3

func (s *OrderService) Checkout(ctx context.Context, userID int, req *domain.CheckoutRequest) (*domain.CheckoutResponse, error) {
	if req.ShippingAddress == "" {
		return nil, ErrShippingAddressEmpty
	}

	if req.Notes == "" {
		req.Notes = "No notes"
	}

	user, err := s.userRepo.FindById(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user for checkout: %w", err)
	}
	if !user.FullName.Valid || user.FullName.String == "" {
		return nil, ErrFullNameRequired
	}
	if !user.Phone.Valid || user.Phone.String == "" {
		return nil, ErrPhoneRequired
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin checkout transaction: %w", err)
	}
	defer tx.Rollback()

	cartItems, err := s.cartRepo.GetCartItemsByUserIDTx(ctx, tx, userID)
	if err != nil {
		return nil, fmt.Errorf("get cart items: %w", err)
	}
	if len(cartItems) == 0 {
		return nil, ErrCartEmpty
	}
	cartID := cartItems[0].CartID

	var totalAmount float64
	orderItems := make([]domain.OrderItem, 0, len(cartItems))

	for _, cartItem := range cartItems {
		variant, err := s.variantRepo.GetVariantByID(ctx, cartItem.VariantID)
		if err != nil {
			if errors.Is(err, repository.ErrVariantNotFound) {
				return nil, fmt.Errorf("%w: variant id %d", ErrVariantNotFound, cartItem.VariantID)
			}
			return nil, fmt.Errorf("get variant: %w", err)
		}

		pricePerItem := cartItem.PriceAtTime
		totalItemPrice := float64(cartItem.Quantity) * pricePerItem
		totalAmount += totalItemPrice

		orderItems = append(orderItems, domain.OrderItem{
			VariantID:    cartItem.VariantID,
			Quantity:     cartItem.Quantity,
			PricePerItem: pricePerItem,
			TotalPrice:   totalItemPrice,
		})

		if err := s.variantRepo.DecreaseStockWithTx(ctx, tx, cartItem.VariantID, cartItem.Quantity); err != nil {
			if errors.Is(err, repository.ErrInsufficientStock) {
				return nil, fmt.Errorf("%w: variant %s", ErrInsufficientStock, variant.VariantName)
			}
			return nil, fmt.Errorf("decrease stock: %w", err)
		}
	}

	shippingCost := 0.0
	tax := totalAmount * taxRate
	grandTotal := totalAmount + shippingCost + tax

	order := &domain.Order{
		UserID:          userID,
		TotalAmount:     totalAmount,
		ShippingCost:    shippingCost,
		Tax:             tax,
		GrandTotal:      grandTotal,
		Status:          "pending",
		ShippingAddress: req.ShippingAddress,
		PaymentMethod:   "bank_transfer",
		Notes:           nullable.ToNullString(req.Notes),
	}

	for attempt := 1; ; attempt++ {
		order.OrderNumber = generateOrderNumber()
		err = s.orderRepo.CreateOrderWithTx(ctx, tx, order)
		if err == nil {
			break
		}
		if errors.Is(err, repository.ErrDuplicateOrderNumber) && attempt < maxOrderNumberAttempts {
			continue
		}
		return nil, fmt.Errorf("create order: %w", err)
	}

	for i := range orderItems {
		orderItems[i].OrderID = order.ID
		if err := s.orderRepo.CreateOrderItemWithTx(ctx, tx, &orderItems[i]); err != nil {
			return nil, fmt.Errorf("create order item: %w", err)
		}
	}

	if err := s.orderRepo.CreateOrderStatusWithTx(ctx, tx, order.ID, "pending", nullable.ToNullString("order created"), "system"); err != nil {
		return nil, fmt.Errorf("create order status: %w", err)
	}

	if err := s.cartRepo.ClearCartWithTx(ctx, tx, cartID); err != nil {
		return nil, fmt.Errorf("clear cart: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit checkout transaction: %w", err)
	}

	order.Items = orderItems
	return &domain.CheckoutResponse{
		Order:      *order,
		Items:      orderItems,
		GrandTotal: grandTotal,
	}, nil
}

func generateOrderNumber() string {
	return fmt.Sprintf("ORD-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID int) ([]domain.Order, error) {
	orders, err := s.orderRepo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user order: %w", err)
	}

	return orders, nil
}

func (s *OrderService) GetOrderDetail(ctx context.Context, userID, orderID int) (*domain.Order, error) {
	order, err := s.orderRepo.GetOrderByID(ctx, userID, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order by id: %w", err)
	}

	return s.enrichOrder(ctx, order), nil
}

func (s *OrderService) GetOrderDetailForAdmin(ctx context.Context, orderID int) (*domain.Order, error) {
	order, err := s.orderRepo.GetOrderByIDForAdmin(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order by id for admin: %w", err)
	}

	return s.enrichOrder(ctx, order), nil
}

func (s *OrderService) enrichOrder(ctx context.Context, order *domain.Order) *domain.Order {
	items, err := s.orderRepo.GetOrderItemsByOrderID(ctx, order.ID)
	if err != nil {
		log.Printf("get items for order %d: %v", order.ID, err)
	} else {
		order.Items = items
	}

	statuses, err := s.orderRepo.GetOrderStatusesByOrderID(ctx, order.ID)
	if err != nil {
		log.Printf("get statuses for order %d: %v", order.ID, err)
	} else {
		order.Statuses = statuses
	}

	return order
}

var orderStatusTransitions = map[string][]string{
	"pending":    {"paid", "canceled"},
	"paid":       {"processing", "canceled", "refunded"},
	"processing": {"shipped", "canceled"},
	"shipped":    {"delivered"},
	"delivered":  {"refunded"},
	"canceled":   {},
	"refunded":   {},
}

var orderStatusTimestampColumn = map[string]string{
	"paid":      "paid_at",
	"shipped":   "shipped_at",
	"delivered": "delivered_at",
}

func isValidStatusTransition(from, to string) bool {
	return slices.Contains(orderStatusTransitions[from], to)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID int, req domain.UpdateOrderStatusReq) error {
	if _, ok := orderStatusTransitions[req.Status]; !ok {
		return fmt.Errorf("%w: %q", ErrInvalidOrderStatus, req.Status)
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin update status transaction: %w", err)
	}
	defer tx.Rollback()

	currentStatus, err := s.orderRepo.GetOrderStatusForUpdate(ctx, tx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return ErrOrderNotFound
		}
		return fmt.Errorf("get current order status: %w", err)
	}

	if !isValidStatusTransition(currentStatus, req.Status) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidStatusTransition, currentStatus, req.Status)
	}

	timestampCol := orderStatusTimestampColumn[req.Status]

	if err = s.orderRepo.UpdateOrderStatusWithTx(ctx, tx, orderID, req.Status, timestampCol); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	if err = s.orderRepo.CreateOrderStatusWithTx(ctx, tx, orderID, req.Status, nullable.ToNullString(req.Notes), "admin"); err != nil {
		return fmt.Errorf("create order status history: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit update status transaction: %w", err)
	}

	return nil
}
