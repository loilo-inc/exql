package exql

import (
	"database/sql"
	"github.com/apex/log"
	"reflect"
	"time"
)

type DB interface {
	Insert(modelPtr interface{}) (sql.Result, error)
	Update(table string, set map[string]interface{}, where Clause) (sql.Result, error)
	Generate(opts *GenerateOptions) error
	DB() *sql.DB
	Close() error
}

type db struct {
	db *sql.DB
	s  Saver
}

func (d *db) Close() error {
	return d.db.Close()
}

func (d *db) DB() *sql.DB {
	return d.db
}

type OpenOptions struct {
	Url string
	// @default 5
	MaxRetryCount *int
	// @default 5s
	RetryInterval *time.Duration
}

func Open(opts *OpenOptions) (DB, error) {
	maxRetryCount := 5
	retryInterval := 5 * time.Second
	if opts.MaxRetryCount != nil {
		maxRetryCount = *opts.MaxRetryCount
	}
	if opts.RetryInterval != nil {
		retryInterval = *opts.RetryInterval
	}
	var d *sql.DB
	var err error
	retryCnt := 0
	for retryCnt < maxRetryCount {
		d, err = sql.Open("mysql", opts.Url)
		if err != nil {
			goto retry
		} else if err = d.Ping(); err != nil {
			goto retry
		} else {
			goto success
		}
	retry:
		log.Errorf("failed to connect database: %s, retrying after %ds...", err, int(retryInterval.Seconds()))
		<-time.NewTimer(retryInterval).C
		retryCnt++
	}
	if err != nil {
		return nil, err
	}
success:
	return &db{
		db: d,
		s:  &saver{},
	}, nil
}

func (d *db) Insert(modelPtr interface{}) (sql.Result, error) {
	s, err := d.s.Insert(modelPtr)
	if err != nil {
		return nil, err
	}
	result, err := d.db.Exec(s.Query, s.Values...)
	if err != nil {
		return nil, err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	kind := s.PrimaryKeyField.Kind()
	if kind == reflect.Int64 {
		s.PrimaryKeyField.Set(reflect.ValueOf(lid))
	} else if kind == reflect.Uint64 {
		s.PrimaryKeyField.Set(reflect.ValueOf(uint64(lid)))
	} else {
		log.Warn("primary key is not int64/uint64. assigning lastInsertedId is skipped")
	}
	return result, nil
}

func (d *db) Update(table string, set map[string]interface{}, where Clause) (sql.Result, error) {
	s, err := d.s.Update(table, set, where)
	if err != nil {
		return nil, err
	}
	return d.db.Exec(s.Query, s.Values...)
}
