package repo

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type SubjectBrand struct {
	Id        uint
	SubjectId uint
	BrandId   uint
}

type SubjectBrandTable struct {
	db               *pgx.Conn
	getAllStmt       *pgconn.StatementDescription
	getBySubjectStmt *pgconn.StatementDescription
	getByBrandStmt   *pgconn.StatementDescription
	insertStmt       *pgconn.StatementDescription
}

const (
	getAllSubjectsBrandsQuery = `SELECT * FROM subjects_brands`
	getBySubjectQuery         = `SELECT * FROM subjects_brands WHERE subject_id = $1`
	getByBrandQuery           = `SELECT * FROM subjects_brands WHERE brand_id = $1`
	insertSubjectBrandQuery   = `INSERT INTO subjects_brands (subject_id, brand_id) values ($1, $2)`
)

func NewSubjectBrandTable(db *pgx.Conn) (*SubjectBrandTable, error) {
	var (
		err              error
		getAllStmt       *pgconn.StatementDescription
		getBySubjectStmt *pgconn.StatementDescription
		getByBrandStmt   *pgconn.StatementDescription
		insertStmt       *pgconn.StatementDescription
	)

	getAllStmt, err = db.Prepare(context.Background(), "getAllSubjectsBrandsQuery", getAllSubjectsBrandsQuery)
	if err != nil {
		return nil, err
	}

	getBySubjectStmt, err = db.Prepare(context.Background(), "getBySubjectQuery", getBySubjectQuery)
	if err != nil {
		return nil, err
	}

	getByBrandStmt, err = db.Prepare(context.Background(), "getByBrandQuery", getByBrandQuery)
	if err != nil {
		return nil, err
	}

	insertStmt, err = db.Prepare(context.Background(), "insertSubjectBrandQuery", insertSubjectBrandQuery)
	if err != nil {
		return nil, err
	}

	return &SubjectBrandTable{
		db:               db,
		getAllStmt:       getAllStmt,
		getBySubjectStmt: getBySubjectStmt,
		getByBrandStmt:   getByBrandStmt,
		insertStmt:       insertStmt,
	}, nil
}

func (t *SubjectBrandTable) GetBrandIdsBySubjectId(subjectId uint) ([]uint, error) {
	rows, err := t.db.Query(context.Background(), t.getBySubjectStmt.Name, subjectId)
	if err != nil {
		return nil, err
	}

	var res []uint
	for rows.Next() {
		var record SubjectBrand

		err = rows.Scan(&record.Id, &record.SubjectId, &record.BrandId)
		if err != nil {
			return nil, err
		}

		res = append(res, record.BrandId)
	}

	rows.Close()

	return res, rows.Err()
}

func (t *SubjectBrandTable) GetSubjectIdsByBrandId(brandId uint) ([]uint, error) {
	rows, err := t.db.Query(context.Background(), t.getBySubjectStmt.Name, brandId)
	if err != nil {
		return nil, err
	}

	var res []uint
	for rows.Next() {
		var record SubjectBrand

		err = rows.Scan(&record.Id, &record.SubjectId, &record.BrandId)
		if err != nil {
			return nil, err
		}

		res = append(res, record.SubjectId)
	}

	rows.Close()

	return res, rows.Err()
}

func (t *SubjectBrandTable) Insert(s SubjectBrand) error {
	_, err := t.db.Exec(context.Background(), t.insertStmt.Name, s.SubjectId, s.BrandId)
	return err
}
