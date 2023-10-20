package repo

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Brand struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

type BrandsTable struct {
	db         *pgx.Conn
	getAllStmt *pgconn.StatementDescription
	insertStmt *pgconn.StatementDescription
	updateStmt *pgconn.StatementDescription
	deleteStmt *pgconn.StatementDescription
}

const (
	getAllBrandsQuery = `SELECT * FROM brands`
	insertBrandQuery  = `INSERT INTO brands (name) values ($1)`
	updateBrandQuery  = `UPDATE brands SET name = $2 WHERE id = $1`
	deleteBrandQuery  = `DELETE FROM brands WHERE id = $1`
)

func NewBrandsTable(db *pgx.Conn) (*BrandsTable, error) {
	var (
		err        error
		getAllStmt *pgconn.StatementDescription
		insertStmt *pgconn.StatementDescription
		updateStmt *pgconn.StatementDescription
		deleteStmt *pgconn.StatementDescription
	)

	getAllStmt, err = db.Prepare(context.Background(), "getAllBrandsQuery", getAllBrandsQuery)
	if err != nil {
		return nil, err
	}

	insertStmt, err = db.Prepare(context.Background(), "insertBrandQuery", insertBrandQuery)
	if err != nil {
		return nil, err
	}

	updateStmt, err = db.Prepare(context.Background(), "updateBrandQuery", updateBrandQuery)
	if err != nil {
		return nil, err
	}

	deleteStmt, err = db.Prepare(context.Background(), "deleteBrandQuery", deleteBrandQuery)
	if err != nil {
		return nil, err
	}

	return &BrandsTable{
		db:         db,
		getAllStmt: getAllStmt,
		insertStmt: insertStmt,
		updateStmt: updateStmt,
		deleteStmt: deleteStmt,
	}, nil
}

func (t *BrandsTable) GetAll() ([]Brand, error) {
	rows, err := t.db.Query(context.Background(), t.getAllStmt.Name)
	if err != nil {
		return nil, err
	}

	var res []Brand
	for rows.Next() {
		var b Brand

		err = rows.Scan(&b.Id, &b.Name)
		if err != nil {
			return nil, err
		}

		res = append(res, b)
	}

	rows.Close()

	return res, rows.Err()
}

func (t *BrandsTable) Insert(s Brand) error {
	_, err := t.db.Exec(context.Background(), t.insertStmt.Name, s.Name)
	return err
}

func (t *BrandsTable) Update(s Brand) error {
	_, err := t.db.Exec(context.Background(), t.updateStmt.Name, s.Id, s.Name)
	return err
}

func (t *BrandsTable) Delete(id uint) error {
	_, err := t.db.Exec(context.Background(), t.deleteStmt.Name, id)
	return err
}
