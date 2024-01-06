package repo

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Subject struct {
	Id    uint   `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

type SubjectsTable struct {
	db *pgxpool.Pool
}

const (
	getAllSubjectsQuery = `SELECT * FROM subjects`
	insertSubjectQuery  = `INSERT INTO subjects (name, image) values ($1, $2)`
	updateSubjectQuery  = `UPDATE subjects SET name = $2, image = $3 WHERE id = $1`
	deleteSubjectQuery  = `DELETE FROM subjects WHERE id = $1`
)

func NewSubjectsTable(db *pgxpool.Pool) *SubjectsTable {
	return &SubjectsTable{db}
}

func (t *SubjectsTable) GetAll() ([]Subject, error) {
	rows, err := t.db.Query(context.Background(), getAllSubjectsQuery)
	if err != nil {
		return nil, err
	}

	var res []Subject
	for rows.Next() {
		var b Subject

		err = rows.Scan(&b.Id, &b.Name, &b.Image)
		if err != nil {
			return nil, err
		}

		res = append(res, b)
	}

	rows.Close()

	return res, rows.Err()
}

func (t *SubjectsTable) Insert(s Subject) error {
	_, err := t.db.Exec(context.Background(), insertSubjectQuery, s.Name, s.Image)
	return err
}

func (t *SubjectsTable) Update(s Subject) error {
	_, err := t.db.Exec(context.Background(), updateSubjectQuery, s.Id, s.Name, s.Image)
	return err
}

func (t *SubjectsTable) Delete(id uint) error {
	_, err := t.db.Exec(context.Background(), deleteSubjectQuery, id)
	return err
}
