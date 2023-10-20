package repo

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Subject struct {
	Id    uint   `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

type SubjectsTable struct {
	db         *pgx.Conn
	getAllStmt *pgconn.StatementDescription
	insertStmt *pgconn.StatementDescription
	updateStmt *pgconn.StatementDescription
	deleteStmt *pgconn.StatementDescription
}

const (
	getAllSubjectsQuery = `SELECT * FROM subjects`
	insertSubjectQuery  = `INSERT INTO subjects (name, image) values ($1, $2)`
	updateSubjectQuery  = `UPDATE subjects SET name = $2, image = $3 WHERE id = $1`
	deleteSubjectQuery  = `DELETE FROM subjects WHERE id = $1`
)

func NewSubjectsTable(db *pgx.Conn) (*SubjectsTable, error) {
	var (
		err        error
		getAllStmt *pgconn.StatementDescription
		insertStmt *pgconn.StatementDescription
		updateStmt *pgconn.StatementDescription
		deleteStmt *pgconn.StatementDescription
	)

	getAllStmt, err = db.Prepare(context.Background(), "getAllSubjectsQuery", getAllSubjectsQuery)
	if err != nil {
		return nil, err
	}

	insertStmt, err = db.Prepare(context.Background(), "insertSubjectQuery", insertSubjectQuery)
	if err != nil {
		return nil, err
	}

	updateStmt, err = db.Prepare(context.Background(), "updateSubjectQuery", updateSubjectQuery)
	if err != nil {
		return nil, err
	}

	deleteStmt, err = db.Prepare(context.Background(), "deleteSubjectQuery", deleteSubjectQuery)
	if err != nil {
		return nil, err
	}

	return &SubjectsTable{
		db:         db,
		getAllStmt: getAllStmt,
		insertStmt: insertStmt,
		updateStmt: updateStmt,
		deleteStmt: deleteStmt,
	}, nil
}

func (t *SubjectsTable) GetAll() ([]Subject, error) {
	rows, err := t.db.Query(context.Background(), t.getAllStmt.Name)
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
	_, err := t.db.Exec(context.Background(), t.insertStmt.Name, s.Name, s.Image)
	return err
}

func (t *SubjectsTable) Update(s Subject) error {
	_, err := t.db.Exec(context.Background(), t.updateStmt.Name, s.Id, s.Name, s.Image)
	return err
}

func (t *SubjectsTable) Delete(id uint) error {
	_, err := t.db.Exec(context.Background(), t.deleteStmt.Name, id)
	return err
}
