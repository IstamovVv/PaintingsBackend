package repo

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Brand struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

type BrandsTable struct {
	db *pgxpool.Pool
}

const (
	getAllBrandsQuery = `SELECT * FROM brands`
	insertBrandQuery  = `INSERT INTO brands (name) values ($1)`
	updateBrandQuery  = `UPDATE brands SET name = $2 WHERE id = $1`
	deleteBrandQuery  = `DELETE FROM brands WHERE id = $1`
)

func NewBrandsTable(db *pgxpool.Pool) *BrandsTable {
	return &BrandsTable{db}
}

func (t *BrandsTable) GetAll() ([]Brand, error) {
	rows, err := t.db.Query(context.Background(), getAllBrandsQuery)
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
	_, err := t.db.Exec(context.Background(), insertBrandQuery, s.Name)
	return err
}

func (t *BrandsTable) Update(s Brand) error {
	_, err := t.db.Exec(context.Background(), updateBrandQuery, s.Id, s.Name)
	return err
}

func (t *BrandsTable) Delete(id uint) error {
	_, err := t.db.Exec(context.Background(), deleteBrandQuery, id)
	return err
}
