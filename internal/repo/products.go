package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	Stock           StockType   `json:"stock"`
	Discount        uint8       `json:"discount"`
	Description     string      `json:"description"`
	Characteristics [][2]string `json:"characteristics"`
	SubjectId       uint        `json:"subject_id"`
	BrandId         uint        `json:"brand_id"`
}

type ProductsTable struct {
	db         *pgx.Conn
	insertStmt *pgconn.StatementDescription
	updateStmt *pgconn.StatementDescription
	deleteStmt *pgconn.StatementDescription
}

const (
	insertProductQuery = `INSERT INTO products (name, stock, price, discount, images, description, characteristics, subject_id, brand_id) values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	updateProductQuery = `UPDATE products SET name = $2, stock = $3, price = $4, discount = $5, images = $6, description = $7, characteristics = $8, subject_id = $9, brand_id = $10 WHERE id = $1`
	deleteProductQuery = `DELETE FROM products WHERE id = $1`
)

func NewProductsTable(db *pgx.Conn) (*ProductsTable, error) {
	var (
		err        error
		insertStmt *pgconn.StatementDescription
		updateStmt *pgconn.StatementDescription
		deleteStmt *pgconn.StatementDescription
	)

	insertStmt, err = db.Prepare(context.Background(), "insertProductQuery", insertProductQuery)
	if err != nil {
		return nil, err
	}

	updateStmt, err = db.Prepare(context.Background(), "updateProductQuery", updateProductQuery)
	if err != nil {
		return nil, err
	}

	deleteStmt, err = db.Prepare(context.Background(), "deleteProductQuery", deleteProductQuery)
	if err != nil {
		return nil, err
	}

	return &ProductsTable{
		db:         db,
		insertStmt: insertStmt,
		updateStmt: updateStmt,
		deleteStmt: deleteStmt,
	}, nil
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
		err = rows.Scan(&p.Id, &p.Name, &p.Stock, &p.Price, &p.Discount, &p.Images, &p.Description, &charBytes, &p.SubjectId, &p.BrandId)
		if err != nil {
			return nil, err
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
		_, err = t.db.Exec(context.Background(), t.updateStmt.Name, p.Id, p.Name, p.Stock, p.Price, p.Discount, p.Images, p.Description, charBytes, p.SubjectId, p.BrandId)
		return err
	}

	_, err = t.db.Exec(context.Background(), t.insertStmt.Name, p.Name, p.Stock, p.Price, p.Discount, p.Images, p.Description, charBytes, p.SubjectId, p.BrandId)
	return err
}

func (t *ProductsTable) Delete(id uint) error {
	_, err := t.db.Exec(context.Background(), t.deleteStmt.Name, id)
	return err
}
