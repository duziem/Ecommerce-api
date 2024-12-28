package order

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/duziem/ecommerce_proj/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
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

func (s *Store) CreateOrder(tx *sql.Tx, order types.Order) (int, error) {
	var orderID int

	// SQL statement to insert a new order into the orders table
	query := `
			INSERT INTO orders (userId, total, status, address, createdAt)
			VALUES ($1, $2, $3, $4, NOW())
			RETURNING id;
	`

	// Execute the query within the transaction
	err := tx.QueryRow(query, order.UserID, order.Total, order.Status, order.Address).Scan(&orderID)
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}

	return orderID, nil
}

func (s *Store) CreateOrderItems(tx *sql.Tx, orderID int, cartItems []types.CartCheckoutItem, products map[int]types.Product) error {
	query := `
			INSERT INTO order_items (orderid, productid, quantity, price)
			VALUES %s;
	`

	var args []interface{}
	var placeholders []string
	for i, item := range cartItems {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
		args = append(args, orderID, item.ProductID, item.Quantity, products[item.ProductID].Price)
	}

	finalQuery := fmt.Sprintf(query, strings.Join(placeholders, ", "))

	if _, err := tx.Exec(finalQuery, args...); err != nil {
		return fmt.Errorf("failed to create order items: %w", err)
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
