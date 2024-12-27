package order

import (
	"database/sql"
	"fmt"

	"github.com/duziem/ecommerce_proj/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateOrder(order types.Order) (int, error) {
	var lastInsertID int
	err := s.db.QueryRow("INSERT INTO orders (userId, total, status, address) VALUES ($1, $2, $3, $4) RETURNING id", order.UserID, order.Total, order.Status, order.Address).Scan(&lastInsertID)
	if err != nil {
		return 0, err
	}

	return lastInsertID, nil
}

func (s *Store) CreateOrderItem(orderItem types.OrderItem) error {
	_, err := s.db.Exec("INSERT INTO order_items (orderId, productId, quantity, price) VALUES ($1, $2, $3, $4)", orderItem.OrderID, orderItem.ProductID, orderItem.Quantity, orderItem.Price)
	return err
}

func (s *Store) GetOrders(userID int) ([]*types.Order, error) {
	rows, err := s.db.Query("SELECT * FROM orders WHERE userId = $1", userID)
	if err != nil {
		return nil, err
	}

	orders := []*types.Order{}
	for rows.Next() {
		order, err := scanRowsIntoOrder(rows)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if len(orders) == 0 {
		return nil, fmt.Errorf("order not found")
	}

	return orders, nil
}

func (s *Store) GetOrderByID(orderID int) (*types.Order, error) {
	rows, err := s.db.Query("SELECT * FROM orders WHERE id = $1", orderID)
	if err != nil {
		return nil, err
	}

	order := new(types.Order)
	for rows.Next() {
		order, err = scanRowsIntoOrder(rows)
		if err != nil {
			return nil, err
		}
	}
	if order.ID == 0 {
		return nil, fmt.Errorf("order not found")
	}

	return order, nil
}

func (s *Store) UpdateOrderStatus(order *types.Order, status string) error {
	query := "UPDATE orders SET status = $1 WHERE id = $2"

	_, err := s.db.Exec(query, status, order.ID)
	if err != nil {
		return err
	}

	return nil
}

func scanRowsIntoOrder(rows *sql.Rows) (*types.Order, error) {
	order := new(types.Order)

	err := rows.Scan(
		&order.ID,
		&order.UserID,
		&order.Total,
		&order.Status,
		&order.Address,
		&order.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return order, nil
}
