package product

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/duziem/ecommerce_proj/types"
	"github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetProductByID(productID int) (*types.Product, error) {
	rows, err := s.db.Query("SELECT * FROM products WHERE id = $1", productID)
	if err != nil {
		return nil, err
	}

	p := new(types.Product)
	for rows.Next() {
		p, err = scanRowsIntoProduct(rows)
		if err != nil {
			return nil, err
		}
	}

	if p.ID == 0 {
		return nil, fmt.Errorf("product not found")
	}

	return p, nil
}

func (s *Store) GetProductsByID(productIDs []int) ([]types.Product, error) {
	numPlaceholders := len(productIDs)
	placeholders := make([]string, numPlaceholders)
	for i := range productIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1) // Create placeholders $1, $2, ...
	}

	query := fmt.Sprintf("SELECT * FROM products WHERE id IN (%s)", strings.Join(placeholders, ","))
	// Convert productIDs to []interface{}
	args := make([]interface{}, len(productIDs))
	for i, v := range productIDs {
		args[i] = v
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	products := []types.Product{}
	for rows.Next() {
		p, err := scanRowsIntoProduct(rows)
		if err != nil {
			return nil, err
		}

		products = append(products, *p)
	}

	return products, nil

}

func (s *Store) DeleteProducts(productIDs []int) error {
	numPlaceholders := len(productIDs)
	placeholders := make([]string, numPlaceholders)
	for i := range productIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1) // Create placeholders $1, $2, ...
	}

	query := fmt.Sprintf("DELETE FROM products WHERE id IN (%s)", strings.Join(placeholders, ","))
	// Convert productIDs to []interface{}
	args := make([]interface{}, len(productIDs))
	for i, v := range productIDs {
		args[i] = v
	}

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetProducts() ([]*types.Product, error) {
	rows, err := s.db.Query("SELECT * FROM products")
	if err != nil {
		return nil, err
	}

	products := make([]*types.Product, 0)
	for rows.Next() {
		p, err := scanRowsIntoProduct(rows)
		if err != nil {
			return nil, err
		}

		products = append(products, p)
	}

	return products, nil
}

func (s *Store) CreateProduct(product types.CreateProductPayload) error {
	_, err := s.db.Exec("INSERT INTO products (name, price, image, description, quantity) VALUES ($1, $2, $3, $4, $5)",
		product.Name, product.Price, product.Image, product.Description, product.Quantity)
	if err != nil {
		return err
	}

	return nil
}

// Store method to update a product in the database
func (s *Store) UpdateProduct(product types.Product) error {
	query := `
	      UPDATE products
	      SET name = COALESCE(NULLIF($1, ''), name),
	          price = COALESCE($2, price),
	          image = COALESCE(NULLIF($3, ''), image),
	          description = COALESCE(NULLIF($4, ''), description),
	          quantity = COALESCE($5, quantity)
	      WHERE id = $6`

	_, err := s.db.Exec(query,
		product.Name,
		product.Price,
		product.Image,
		product.Description,
		product.Quantity,
		product.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteProduct(productID int) error {
	_, err := s.db.Exec("DELETE FROM products WHERE id = $1", productID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) BeginTransaction() (*sql.Tx, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

func (s *Store) GetProductsByIDWithLock(tx *sql.Tx, ids []int) ([]types.Product, error) {
	query := `
			SELECT * 
			FROM products 
			WHERE id = ANY($1) 
			FOR UPDATE;
	`

	rows, err := tx.Query(query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}
	defer rows.Close()

	var products []types.Product
	for rows.Next() {
		product, err := scanRowsIntoProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, *product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return products, nil
}

func (s *Store) UpdateProductQuantities(tx *sql.Tx, cartItems []types.CartCheckoutItem) error {
	query := `
		UPDATE products
		SET quantity = quantity - excluded.product_quantity::integer
		FROM (VALUES %s) AS excluded(product_id, product_quantity)
		WHERE products.id = excluded.product_id::integer;
	`

	// Build dynamic VALUES clause
	var args []interface{}
	var placeholders []string
	for i, item := range cartItems {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		args = append(args, item.ProductID, item.Quantity)
	}

	finalQuery := fmt.Sprintf(query, strings.Join(placeholders, ", "))

	if _, err := tx.Exec(finalQuery, args...); err != nil {
		return fmt.Errorf("failed to update product quantities: %w", err)
	}

	return nil
}

func scanRowsIntoProduct(rows *sql.Rows) (*types.Product, error) {
	product := new(types.Product)

	err := rows.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Image,
		&product.Price,
		&product.Quantity,
		&product.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return product, nil
}
