package repo

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
}

type ProductsTable struct {
	db         *pgx.Conn
	getAllStmt *pgconn.StatementDescription
	insertStmt *pgconn.StatementDescription
	updateStmt *pgconn.StatementDescription
	deleteStmt *pgconn.StatementDescription
}

const (
	getAllQuery = `SELECT * FROM products OFFSET $1 LIMIT $2`
	insertQuery = `INSERT INTO products (name, stock, price, discount, images, description, characteristics) values ($1, $2, $3, $4, $5, $6, $7)`
	updateQuery = `UPDATE products SET name = $2, stock = $3, price = $4, discount = $5, images = $6, description = $7, characteristics = $8 WHERE id = $1`
	deleteQuery = `DELETE FROM products WHERE id = $1`
)

func NewProductsTable(db *pgx.Conn) (*ProductsTable, error) {
	var (
		err        error
		getAllStmt *pgconn.StatementDescription
		insertStmt *pgconn.StatementDescription
		updateStmt *pgconn.StatementDescription
		deleteStmt *pgconn.StatementDescription
	)

	getAllStmt, err = db.Prepare(context.Background(), "getAllProductsQuery", getAllQuery)
	if err != nil {
		return nil, err
	}

	insertStmt, err = db.Prepare(context.Background(), "insertProductQuery", insertQuery)
	if err != nil {
		return nil, err
	}

	updateStmt, err = db.Prepare(context.Background(), "updateProductQuery", updateQuery)
	if err != nil {
		return nil, err
	}

	deleteStmt, err = db.Prepare(context.Background(), "deleteProductQuery", deleteQuery)
	if err != nil {
		return nil, err
	}

	return &ProductsTable{
		db:         db,
		getAllStmt: getAllStmt,
		insertStmt: insertStmt,
		updateStmt: updateStmt,
		deleteStmt: deleteStmt,
	}, nil
}

func (t *ProductsTable) GetAllProducts(offset int, limit int) ([]Product, error) {
	rows, err := t.db.Query(context.Background(), t.getAllStmt.Name, offset, limit)
	if err != nil {
		return nil, err
	}

	var res []Product
	for rows.Next() {
		var p Product

		var charBytes []byte
		err = rows.Scan(&p.Id, &p.Name, &p.Stock, &p.Price, &p.Discount, &p.Images, &p.Description, &charBytes)
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

func (t *ProductsTable) Insert(p Product, editFlag bool) error {
	charBytes, err := json.Marshal(p.Characteristics)
	if err != nil {
		return err
	}

	if editFlag {
		_, err = t.db.Exec(context.Background(), t.updateStmt.Name, p.Id, p.Name, p.Stock, p.Price, p.Discount, p.Images, p.Description, charBytes)
		return err
	}

	_, err = t.db.Exec(context.Background(), t.insertStmt.Name, p.Name, p.Stock, p.Price, p.Discount, p.Images, p.Description, charBytes)
	return err
}

func (t *ProductsTable) Delete(id uint) error {
	_, err := t.db.Exec(context.Background(), t.deleteStmt.Name, id)
	return err
}
