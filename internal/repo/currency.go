package repo

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Currency struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

type CurrencyTable struct {
	db *pgxpool.Pool
}

const (
	getCurrencyAllQuery = `SELECT * FROM currency`
	insertCurrencyQuery = `INSERT INTO currency (name) values ($1)`
	updateCurrencyQuery = `UPDATE currency SET name = $2 WHERE id = $1`
	deleteCurrencyQuery = `DELETE FROM currency WHERE id = $1`
)

func NewCurrencyTable(db *pgxpool.Pool) *CurrencyTable {
	return &CurrencyTable{db}
}

func (t *CurrencyTable) GetAllCurrency() ([]Currency, error) {
	rows, err := t.db.Query(context.Background(), getCurrencyAllQuery)
	if err != nil {
		return nil, err
	}

	var res []Currency
	for rows.Next() {
		var c Currency

		err = rows.Scan(&c.Id, &c.Name)
		if err != nil {
			return nil, err
		}

		res = append(res, c)
	}

	rows.Close()

	return res, rows.Err()
}

func (t *CurrencyTable) Insert(c Currency, editFlag bool) error {
	var err error

	if editFlag {
		_, err = t.db.Exec(context.Background(), updateCurrencyQuery, c.Id, c.Name)
		return err
	}

	_, err = t.db.Exec(context.Background(), insertCurrencyQuery, c.Name)
	return err
}

func (t *CurrencyTable) Delete(id uint) error {
	_, err := t.db.Exec(context.Background(), deleteCurrencyQuery, id)
	return err
}
