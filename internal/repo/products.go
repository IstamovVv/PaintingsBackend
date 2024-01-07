package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

type StockType int

const (
	OnOrder StockType = iota
	InStock
	OutOfStock
)

type Product struct {
	Id              uint        `json:"id"`
	Name            string      `json:"name"`
	Images          []string    `json:"images"`
	Price           float32     `json:"price"`
	Currency        uint        `json:"currency"`
	Stock           StockType   `json:"stock"`
	Discount        uint8       `json:"discount"`
	Description     string      `json:"description"`
	Characteristics [][2]string `json:"characteristics"`
	SubjectId       uint        `json:"subject"`
	BrandId         uint        `json:"brand"`
}

type ProductsTable struct {
	db *pgxpool.Pool
}

const (
	insertProductQuery = `INSERT INTO products (name, stock, price, currency, discount, images, description, characteristics, subject_id, brand_id) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	updateProductQuery = `UPDATE products SET name = $2, stock = $3, price = $4, currency = $5, discount = $6, images = $7, description = $8, characteristics = $9, subject_id = $10, brand_id = $11 WHERE id = $1`
	deleteProductQuery = `DELETE FROM products WHERE id = $1`
)

func NewProductsTable(db *pgxpool.Pool) *ProductsTable {
	return &ProductsTable{db}
}

type SearchProductsOptions struct {
	Brand       string
	BrandFilter bool

	Subject       string
	SubjectFilter bool
}

func (t *ProductsTable) GetAllProducts(offset int, limit int, options SearchProductsOptions) ([]Product, error) {
	query := t.prepareGetAllQuery(options)
	rows, err := t.db.Query(context.Background(), query, offset, limit)
	if err != nil {
		return nil, err
	}

	var res []Product
	for rows.Next() {
		var p Product

		var charBytes []byte
		var currencyId *uint
		err = rows.Scan(&p.Id, &p.Name, &p.Stock, &p.Price, &p.Discount, &p.Images, &p.Description, &charBytes, &p.SubjectId, &p.BrandId, &currencyId)
		if err != nil {
			return nil, err
		}

		if currencyId != nil {
			p.Currency = *currencyId
		}

		err = json.Unmarshal(charBytes, &p.Characteristics)
		if err != nil {
			return nil, err
		}

		res = append(res, p)
	}
	rows.Close()

	return res, rows.Err()
}

func (t *ProductsTable) prepareGetAllQuery(options SearchProductsOptions) string {
	var builder strings.Builder
	builder.WriteString("SELECT * FROM products")
	var conditions []string
	if options.SubjectFilter {
		conditions = append(conditions, fmt.Sprintf("subject_id = %s", options.Subject))
	}
	if options.BrandFilter {
		conditions = append(conditions, fmt.Sprintf("brand_id = %s", options.Brand))
	}

	conditionsStr := strings.Join(conditions, " AND ")
	if len(conditionsStr) != 0 {
		builder.WriteString(" WHERE " + conditionsStr)
	}
	builder.WriteString(" OFFSET $1 LIMIT $2")

	return builder.String()
}

func (t *ProductsTable) Insert(p Product, editFlag bool) error {
	charBytes, err := json.Marshal(p.Characteristics)
	if err != nil {
		return err
	}

	if editFlag {
		_, err = t.db.Exec(context.Background(), updateProductQuery, p.Id, p.Name, p.Stock, p.Price, p.Currency, p.Discount, p.Images, p.Description, charBytes, p.SubjectId, p.BrandId)
		return err
	}

	_, err = t.db.Exec(context.Background(), insertProductQuery, p.Name, p.Stock, p.Price, p.Currency, p.Discount, p.Images, p.Description, charBytes, p.SubjectId, p.BrandId)
	return err
}

func (t *ProductsTable) Delete(id uint) error {
	_, err := t.db.Exec(context.Background(), deleteProductQuery, id)
	return err
}
