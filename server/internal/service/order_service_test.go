package service_test

import (
	"context"
	"errors"
	"furniture-api/internal/domain"
	"furniture-api/internal/nullable"
	"furniture-api/internal/service"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

// --- order repo mock ---

type mockOrderRepo struct {
	getOrdersByUserIDFn        func(ctx context.Context, userID int) ([]domain.Order, error)
	getOrderByIDFn             func(ctx context.Context, userID, orderID int) (*domain.Order, error)
	getOrderByIDForAdminFn     func(ctx context.Context, orderID int) (*domain.Order, error)
	getOrderItemsByOrderIDFn   func(ctx context.Context, orderID int) ([]domain.OrderItem, error)
	getOrderStatusesByOrderIDFn func(ctx context.Context, orderID int) ([]domain.OrderStatus, error)
	getOrderSummariesFn        func(ctx context.Context, orderIDs []int) (map[int]domain.OrderSummary, error)
}

func (m *mockOrderRepo) CreateOrderWithTx(ctx context.Context, tx *sqlx.Tx, order *domain.Order) error {
	panic("not implemented")
}
func (m *mockOrderRepo) CreateOrderItemWithTx(ctx context.Context, tx *sqlx.Tx, item *domain.OrderItem) error {
	panic("not implemented")
}
func (m *mockOrderRepo) CreateOrderStatusWithTx(ctx context.Context, tx *sqlx.Tx, orderID int, status string, notes nullable.NullString, createdBy string) error {
	panic("not implemented")
}
func (m *mockOrderRepo) UpdateOrderStatusWithTx(ctx context.Context, tx *sqlx.Tx, orderID int, status string, timestampColumn string) error {
	panic("not implemented")
}
func (m *mockOrderRepo) GetOrdersByUserID(ctx context.Context, userID int) ([]domain.Order, error) {
	return m.getOrdersByUserIDFn(ctx, userID)
}
func (m *mockOrderRepo) GetOrderByID(ctx context.Context, userID, orderID int) (*domain.Order, error) {
	return m.getOrderByIDFn(ctx, userID, orderID)
}
func (m *mockOrderRepo) GetOrderByIDForAdmin(ctx context.Context, orderID int) (*domain.Order, error) {
	return m.getOrderByIDForAdminFn(ctx, orderID)
}
func (m *mockOrderRepo) GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]domain.OrderItem, error) {
	return m.getOrderItemsByOrderIDFn(ctx, orderID)
}
func (m *mockOrderRepo) GetOrderStatusesByOrderID(ctx context.Context, orderID int) ([]domain.OrderStatus, error) {
	return m.getOrderStatusesByOrderIDFn(ctx, orderID)
}
func (m *mockOrderRepo) GetOrderStatusForUpdate(ctx context.Context, tx *sqlx.Tx, orderID int) (string, error) {
	panic("not implemented")
}
func (m *mockOrderRepo) GetOrderSummaries(ctx context.Context, orderIDs []int) (map[int]domain.OrderSummary, error) {
	return m.getOrderSummariesFn(ctx, orderIDs)
}

// --- helpers ---

func dummyOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{
		getOrdersByUserIDFn: func(ctx context.Context, userID int) ([]domain.Order, error) {
			return nil, errors.New("GetOrdersByUserID: unexpected call")
		},
		getOrderByIDFn: func(ctx context.Context, userID, orderID int) (*domain.Order, error) {
			return nil, errors.New("GetOrderByID: unexpected call")
		},
		getOrderByIDForAdminFn: func(ctx context.Context, orderID int) (*domain.Order, error) {
			return nil, errors.New("GetOrderByIDForAdmin: unexpected call")
		},
		getOrderItemsByOrderIDFn: func(ctx context.Context, orderID int) ([]domain.OrderItem, error) {
			return nil, errors.New("GetOrderItemsByOrderID: unexpected call")
		},
		getOrderStatusesByOrderIDFn: func(ctx context.Context, orderID int) ([]domain.OrderStatus, error) {
			return nil, errors.New("GetOrderStatusesByOrderID: unexpected call")
		},
		getOrderSummariesFn: func(ctx context.Context, orderIDs []int) (map[int]domain.OrderSummary, error) {
			return nil, errors.New("GetOrderSummaries: unexpected call")
		},
	}
}

func newTestOrderService(orderRepo *mockOrderRepo) *service.OrderService {
	return service.NewOrderService(orderRepo, nil, nil, nil, nil, nil)
}

func sampleOrder(id int, userID int, status string) domain.Order {
	return domain.Order{
		ID:          id,
		UserID:      userID,
		OrderNumber: "ORD-20260101-0001",
		TotalAmount: 100000,
		GrandTotal:  112000,
		Status:      status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// --- GetUserOrders ---

func TestGetUserOrders_Success(t *testing.T) {
	repo := dummyOrderRepo()
	repo.getOrdersByUserIDFn = func(ctx context.Context, userID int) ([]domain.Order, error) {
		return []domain.Order{
			sampleOrder(1, userID, "pending"),
			sampleOrder(2, userID, "paid"),
		}, nil
	}
	repo.getOrderSummariesFn = func(ctx context.Context, orderIDs []int) (map[int]domain.OrderSummary, error) {
		return map[int]domain.OrderSummary{
			1: {OrderID: 1, TotalItems: 2, VariantName: "Table", ImageURL: "img1.jpg"},
			2: {OrderID: 2, TotalItems: 1, VariantName: "Chair", ImageURL: "img2.jpg"},
		}, nil
	}

	svc := newTestOrderService(repo)
	orders, err := svc.GetUserOrders(context.Background(), 1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(orders))
	}
	if orders[0].FirstItemName != "Table" {
		t.Errorf("expected first item 'Table', got %q", orders[0].FirstItemName)
	}
	if orders[1].TotalItems != 1 {
		t.Errorf("expected total items 1, got %d", orders[1].TotalItems)
	}
}

func TestGetUserOrders_Empty(t *testing.T) {
	repo := dummyOrderRepo()
	repo.getOrdersByUserIDFn = func(ctx context.Context, userID int) ([]domain.Order, error) {
		return []domain.Order{}, nil
	}

	svc := newTestOrderService(repo)
	orders, err := svc.GetUserOrders(context.Background(), 1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(orders) != 0 {
		t.Fatalf("expected empty slice, got %d orders", len(orders))
	}
}

// --- GetOrderDetail ---

func TestGetOrderDetail_Success(t *testing.T) {
	repo := dummyOrderRepo()
	repo.getOrderByIDFn = func(ctx context.Context, userID, orderID int) (*domain.Order, error) {
		o := sampleOrder(orderID, userID, "processing")
		return &o, nil
	}
	repo.getOrderItemsByOrderIDFn = func(ctx context.Context, orderID int) ([]domain.OrderItem, error) {
		return []domain.OrderItem{{ID: 1, OrderID: orderID, Quantity: 2, PricePerItem: 50000, TotalPrice: 100000}}, nil
	}
	repo.getOrderStatusesByOrderIDFn = func(ctx context.Context, orderID int) ([]domain.OrderStatus, error) {
		return []domain.OrderStatus{{ID: 1, OrderID: orderID, Status: "pending"}, {ID: 2, OrderID: orderID, Status: "processing"}}, nil
	}

	svc := newTestOrderService(repo)
	order, err := svc.GetOrderDetail(context.Background(), 1, 10)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(order.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(order.Items))
	}
	if len(order.Statuses) != 2 {
		t.Errorf("expected 2 statuses, got %d", len(order.Statuses))
	}
}

func TestGetOrderDetail_NotFound(t *testing.T) {
	repo := dummyOrderRepo()
	repo.getOrderByIDFn = func(ctx context.Context, userID, orderID int) (*domain.Order, error) {
		return nil, domain.ErrOrderNotFound
	}

	svc := newTestOrderService(repo)
	_, err := svc.GetOrderDetail(context.Background(), 1, 999)

	if !errors.Is(err, domain.ErrOrderNotFound) {
		t.Fatalf("expected ErrOrderNotFound, got %v", err)
	}
}

// --- GetOrderDetailForAdmin ---

func TestGetOrderDetailForAdmin_Success(t *testing.T) {
	repo := dummyOrderRepo()
	repo.getOrderByIDForAdminFn = func(ctx context.Context, orderID int) (*domain.Order, error) {
		o := sampleOrder(orderID, 1, "shipped")
		return &o, nil
	}
	repo.getOrderItemsByOrderIDFn = func(ctx context.Context, orderID int) ([]domain.OrderItem, error) {
		return []domain.OrderItem{}, nil
	}
	repo.getOrderStatusesByOrderIDFn = func(ctx context.Context, orderID int) ([]domain.OrderStatus, error) {
		return []domain.OrderStatus{}, nil
	}

	svc := newTestOrderService(repo)
	order, err := svc.GetOrderDetailForAdmin(context.Background(), 10)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if order.ID != 10 {
		t.Errorf("expected order ID 10, got %d", order.ID)
	}
}

func TestGetOrderDetailForAdmin_NotFound(t *testing.T) {
	repo := dummyOrderRepo()
	repo.getOrderByIDForAdminFn = func(ctx context.Context, orderID int) (*domain.Order, error) {
		return nil, domain.ErrOrderNotFound
	}

	svc := newTestOrderService(repo)
	_, err := svc.GetOrderDetailForAdmin(context.Background(), 999)

	if !errors.Is(err, domain.ErrOrderNotFound) {
		t.Fatalf("expected ErrOrderNotFound, got %v", err)
	}
}
