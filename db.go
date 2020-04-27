package exql

import (
	"database/sql"
	"github.com/apex/log"
	"time"
)

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
			log.Errorf("failed to connect database: %s, retrying after %ds...", err, int(retryInterval.Seconds()))
			time.Sleep(retryInterval)
		} else {
			break
		}
		retryCnt++
	}
	if err != nil {
		log.Fatalf("failed to connect database: %s", err)
	}
	return &db{db: d}, nil
}
