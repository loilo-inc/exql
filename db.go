package exql

import "database/sql"

type DB interface {
	Insert(modelPtr interface{}) (sql.Result, error)
	Update(table string, set map[string]interface{}, where WhereQuery) (sql.Result, error)
	Generate(opts *GenerateOptions) error
	DB() *sql.DB
}

type db struct {
	db *sql.DB
}

func (d *db) DB() *sql.DB {
	return d.db
}

func Open(url string) (DB, error) {
	d, err := sql.Open("mysql", url)
	if err != nil {
		return nil, err
	}
	return &db{db: d}, nil
}
