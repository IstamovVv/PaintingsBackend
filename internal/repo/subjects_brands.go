package repo

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubjectBrand struct {
	Id        uint
	SubjectId uint
	BrandId   uint
}

type SubjectBrandTable struct {
	db *pgxpool.Pool
}

const (
	getAllSubjectsBrandsQuery = `SELECT * FROM subjects_brands`
	getBySubjectQuery         = `SELECT * FROM subjects_brands WHERE subject_id = $1`
	getByBrandQuery           = `SELECT * FROM subjects_brands WHERE brand_id = $1`
	insertSubjectBrandQuery   = `INSERT INTO subjects_brands (subject_id, brand_id) values ($1, $2)`
)

func NewSubjectBrandTable(db *pgxpool.Pool) *SubjectBrandTable {
	return &SubjectBrandTable{db}
}

func (t *SubjectBrandTable) GetBrandIdsBySubjectId(subjectId uint) ([]uint, error) {
	rows, err := t.db.Query(context.Background(), getBySubjectQuery, subjectId)
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
	rows, err := t.db.Query(context.Background(), getByBrandQuery, brandId)
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
	_, err := t.db.Exec(context.Background(), insertSubjectBrandQuery, s.SubjectId, s.BrandId)
	return err
}
