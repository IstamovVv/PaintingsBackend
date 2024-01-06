package repo

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Subject struct {
	Id       uint   `json:"id"`
	Name     string `json:"name"`
	Image    string `json:"image"`
	ParentId uint   `json:"parentId"`
}

type SubjectV2 struct {
	Id       uint   `json:"id"`
	Name     string `json:"name"`
	Image    string `json:"image"`
	ParentId uint   `json:"parentId"`
	Children []uint `json:"children"`
}

type SubjectsTable struct {
	db *pgxpool.Pool
}

const (
	getAllSubjectsQuery = `SELECT * FROM subjects`
	insertSubjectQuery  = `INSERT INTO subjects (name, image, parent_id) values ($1, $2, $3)`
	updateSubjectQuery  = `UPDATE subjects SET name = $2, image = $3, parent_id = $4 WHERE id = $1`
	deleteSubjectQuery  = `DELETE FROM subjects WHERE id = $1`

	getAllSubjectsQueryV2 = `select s1.id, s1.name, s1.image, s1.parent_id, ARRAY_REMOVE(ARRAY_AGG(s2.id), NULL) children
							 from subjects s1
							 left join subjects s2 on s1.id = s2.parent_id
							 group by s1.id;`
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

		var parentId *uint
		err = rows.Scan(&b.Id, &b.Name, &b.Image, &parentId)
		if err != nil {
			return nil, err
		}

		if parentId != nil {
			b.ParentId = *parentId
		}

		res = append(res, b)
	}

	rows.Close()

	return res, rows.Err()
}

func (t *SubjectsTable) GetAllV2() ([]SubjectV2, error) {
	rows, err := t.db.Query(context.Background(), getAllSubjectsQueryV2)
	if err != nil {
		return nil, err
	}

	var res []SubjectV2
	for rows.Next() {
		var b SubjectV2

		var parentId *uint
		var children pgtype.Array[uint]
		err = rows.Scan(&b.Id, &b.Name, &b.Image, &parentId, &children)
		if err != nil {
			return nil, err
		}

		b.Children = children.Elements

		if parentId != nil {
			b.ParentId = *parentId
		}

		res = append(res, b)
	}

	rows.Close()

	return res, rows.Err()
}

func (t *SubjectsTable) Insert(s Subject) error {
	_, err := t.db.Exec(context.Background(), insertSubjectQuery, s.Name, s.Image, s.ParentId)
	return err
}

func (t *SubjectsTable) Update(s Subject) error {
	_, err := t.db.Exec(context.Background(), updateSubjectQuery, s.Id, s.Name, s.Image, s.ParentId)
	return err
}

func (t *SubjectsTable) Delete(id uint) error {
	_, err := t.db.Exec(context.Background(), deleteSubjectQuery, id)
	return err
}
