package repo

import (
	"context"
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
	Id              uint
	Name            string
	Images          []string
	Stock           StockType
	Discount        uint8
	Description     string
	Characteristics [][2]string
}

type ProductsTable struct {
	db         *pgx.Conn
	getAllStmt *pgconn.StatementDescription
	insertStmt *pgconn.StatementDescription
	deleteStmt *pgconn.StatementDescription
}

const (
	getAllQuery = `SELECT * FROM products OFFSET $1 LIMIT $2`
	insertQuery = `INSERT INTO products (name, images, stock, discount, description, characteristics) values ($1, $2, $3, $4, $5, $6)`
	deleteQuery = `DELETE FROM products WHERE id = $1`
)

func NewProductsTable(db *pgx.Conn) (*ProductsTable, error) {
	var (
		err        error
		getAllStmt *pgconn.StatementDescription
		insertStmt *pgconn.StatementDescription
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

	deleteStmt, err = db.Prepare(context.Background(), "deleteProductQuery", deleteQuery)
	if err != nil {
		return nil, err
	}

	return &ProductsTable{
		db:         db,
		getAllStmt: getAllStmt,
		insertStmt: insertStmt,
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
		var (
			p Product
		)

		err = rows.Scan(&p.Id, &p.Name, &p.Images, &p.Stock, &p.Discount, &p.Description, &p.Characteristics)
		if err != nil {
			return nil, err
		}

		res = append(res, p)
	}
	rows.Close()

	return res, rows.Err()
}

func (t *ProductsTable) Insert(p Product) error {
	_, err := t.db.Exec(context.Background(), t.insertStmt.Name, p.Name, p.Images, p.Stock, p.Discount, p.Description, p.Characteristics)
	return err
}

func (t *ProductsTable) Delete(id uint) error {
	_, err := t.db.Exec(context.Background(), t.deleteStmt.Name, id)
	return err
}
