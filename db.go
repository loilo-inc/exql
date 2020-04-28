package exql

import (
	"database/sql"
	"github.com/apex/log"
	"time"
)

type DB interface {
	Saver
	Mapper
	DB() *sql.DB
	Close() error
}

type db struct {
	db *sql.DB
	s  Saver
	m  Mapper
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
		s:  NewSaver(d),
		m:  NewMapper(),
	}, nil
}

func (d *db) Insert(structPtr interface{}) (sql.Result, error) {
	return d.s.Insert(structPtr)
}

func (d *db) QueryForInsert(structPtr interface{}) (*SaveQuery, error) {
	return d.s.QueryForInsert(structPtr)
}

func (d *db) Update(table string, set map[string]interface{}, where Clause) (sql.Result, error) {
	return d.s.Update(table, set, where)
}

func (d *db) QueryForUpdate(table string, set map[string]interface{}, where Clause) (*SaveQuery, error) {
	return d.s.QueryForUpdate(table, set, where)
}

func (d *db) Map(rows *sql.Rows, pointerOfStruct interface{}) error {
	return d.m.Map(rows, pointerOfStruct)
}

func (d *db) MapMany(rows *sql.Rows, pointerOfSliceOfStruct interface{}) error {
	return d.m.MapMany(rows, pointerOfSliceOfStruct)
}

func (d *db) Close() error {
	return d.db.Close()
}

func (d *db) DB() *sql.DB {
	return d.db
}
