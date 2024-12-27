package product

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
