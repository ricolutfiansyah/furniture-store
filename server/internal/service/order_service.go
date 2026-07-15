package service

import (
	"context"
	"errors"
	"fmt"
	"furniture-api/internal/domain"
	"furniture-api/internal/nullable"
	"furniture-api/internal/repository"
	"furniture-api/internal/validation"
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
	GetOrderSummaries(ctx context.Context, orderIDs []int) (map[int]repository.OrderSummaryDTO, error)
}

type CartRepositoryForOrder interface {
	GetCartItemsByUserIDTx(ctx context.Context, tx *sqlx.Tx, userID int) ([]domain.CartItem, error)
	GetCartItemsByIDsTx(ctx context.Context, tx *sqlx.Tx, userID int, itemIDs []int) ([]domain.CartItem, error)
	RemoveCartItemsWithTx(ctx context.Context, tx *sqlx.Tx, cartID int, itemIDs []int) error
}

type ProductVariantRepositoryForOrder interface {
	GetVariantByID(ctx context.Context, id int) (*domain.ProductVariant, error)
	DecreaseStockWithTx(ctx context.Context, tx *sqlx.Tx, variantID, quantity int) error
}

type UserRepositoryForOrder interface {
	FindById(ctx context.Context, id int) (*domain.User, error)
}

type AddressRepositoryForOrder interface {
	GetByID(ctx context.Context, id, userID int) (*domain.UserAddress, error)
}

type OrderService struct {
	orderRepo   OrderRepository
	cartRepo    CartRepositoryForOrder
	variantRepo ProductVariantRepositoryForOrder
	userRepo    UserRepositoryForOrder
	addressRepo AddressRepositoryForOrder
	db          *sqlx.DB
}

func NewOrderService(
	or OrderRepository,
	cr CartRepositoryForOrder,
	vr ProductVariantRepositoryForOrder,
	ur UserRepositoryForOrder,
	ar AddressRepositoryForOrder,
	db *sqlx.DB,
) *OrderService {
	return &OrderService{
		orderRepo:   or,
		cartRepo:    cr,
		variantRepo: vr,
		userRepo:    ur,
		addressRepo: ar,
		db:          db,
	}
}

const taxRate = 0.12
const maxOrderNumberAttempts = 3

func (s *OrderService) Checkout(ctx context.Context, userID int, req *domain.CheckoutRequest) (*domain.CheckoutResponse, error) {
	if err := validation.Validate(
		validation.ValidateAddressID(req.AddressID),
	); err != nil {
		return nil, err
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

	address, err := s.addressRepo.GetByID(ctx, req.AddressID, userID)
	if err != nil {
		return nil, err
	}
	shippingAddress := formatShippingAddress(address)

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin checkout transaction: %w", err)
	}
	defer tx.Rollback()

	if len(req.CartItemIDs) == 0 {
		return nil, errors.New("Please choose an item first")
	}

	cartItems, err := s.cartRepo.GetCartItemsByIDsTx(ctx, tx, userID, req.CartItemIDs)
	if err != nil {
		return nil, fmt.Errorf("get cart items: %w", err)
	}
	if len(cartItems) == 0 {
		return nil, ErrCartEmpty
	}

	if len(cartItems) != len(req.CartItemIDs) {
		return nil, errors.New("One of the item is invalid or no longer exist in the cart")
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
		ShippingAddress: shippingAddress,
		PaymentMethod:   "bank_transfer",
		Notes:           nullable.NewNullString(req.Notes),
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

	if err := s.orderRepo.CreateOrderStatusWithTx(ctx, tx, order.ID, "pending", nullable.NewNullString("order created"), "system"); err != nil {
		return nil, fmt.Errorf("create order status: %w", err)
	}

	if err := s.cartRepo.RemoveCartItemsWithTx(ctx, tx, cartID, req.CartItemIDs); err != nil {
		return nil, fmt.Errorf("remove checked out items from cart: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit checkout transaction: %w", err)
	}

	order.Items = orderItems
	return &domain.CheckoutResponse{
		Order: *order,
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
	if len(orders) == 0 {
		return orders, nil
	}

	var orderIDs []int
	for _, o := range orders {
		orderIDs = append(orderIDs, o.ID)
	}

	summaries, err := s.orderRepo.GetOrderSummaries(ctx, orderIDs)
	if err != nil {
		log.Printf("failed to get summaries: %v", err)
	}

	for i, order := range orders {
		if summary, exists := summaries[order.ID]; exists {
			orders[i].FirstItemName = summary.VariantName
			orders[i].FirstItemImage = summary.ImageURL
			orders[i].TotalItems = summary.TotalItems
		}
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

	if err = s.orderRepo.CreateOrderStatusWithTx(ctx, tx, orderID, req.Status, nullable.NewNullString(req.Notes), "admin"); err != nil {
		return fmt.Errorf("create order status history: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit update status transaction: %w", err)
	}

	return nil
}
