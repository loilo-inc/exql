package exql

import (
	"context"
	"database/sql"
	"github.com/apex/log"
	"sync"
	"time"
)

type DB interface {
	Saver
	Mapper
	// Return *sql.DB instance
	DB() *sql.DB
	// Set db object
	SetDB(db *sql.DB)
	// Begin transaction and commit.
	// If error returned from callback, transaction is rolled back.
	// Internally call tx.BeginTx(context.Background(), nil)
	Transaction(callback func(tx Tx) error) error
	// Same as Transaction()
	// Internally call tx.BeginTx(ctx, opts)
	TransactionWithContext(ctx context.Context, opts *sql.TxOptions, callback func(tx Tx) error) error
	// Call db.Close()
	Close() error
}

type db struct {
	db    *sql.DB
	s     *saver
	m     *mapper
	mutex sync.Mutex
}

type OpenOptions struct {
	// @default "mysql"
	DriverName string
	Url        string
	// @default 5
	MaxRetryCount int
	// @default 5s
	RetryInterval time.Duration
}

func Open(opts *OpenOptions) (DB, error) {
	driverName := "mysql"
	if opts.DriverName != "" {
		driverName = opts.DriverName
	}
	maxRetryCount := 5
	retryInterval := 5 * time.Second
	if opts.MaxRetryCount > 0 {
		maxRetryCount = opts.MaxRetryCount
	}
	if opts.RetryInterval > 0 {
		retryInterval = opts.RetryInterval
	}
	var d *sql.DB
	var err error
	retryCnt := 0
	for retryCnt < maxRetryCount {
		d, err = sql.Open(driverName, opts.Url)
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
	return NewDB(d), nil
}

func NewDB(d *sql.DB) DB {
	return &db{
		db: d,
		s:  &saver{ex: d},
		m:  &mapper{},
	}
}

func (d *db) Insert(structPtr interface{}) (sql.Result, error) {
	return d.s.Insert(structPtr)
}

func (d *db) InsertContext(ctx context.Context, structPtr interface{}) (sql.Result, error) {
	return d.s.InsertContext(ctx, structPtr)
}

func (d *db) QueryForInsert(structPtr interface{}) (*SaveQuery, error) {
	return d.s.QueryForInsert(structPtr)
}

func (d *db) Update(table string, set map[string]interface{}, where Clause) (sql.Result, error) {
	return d.s.Update(table, set, where)
}

func (d *db) UpdateContext(ctx context.Context, table string, set map[string]interface{}, where Clause) (sql.Result, error) {
	return d.s.UpdateContext(ctx, table, set, where)
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
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.db.Close()
}

func (d *db) DB() *sql.DB {
	return d.db
}

func (d *db) SetDB(db *sql.DB) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.db = db
	d.s.ex = db
}

func (d *db) Transaction(callback func(tx Tx) error) error {
	return d.TransactionWithContext(context.Background(), nil, callback)
}

func (d *db) TransactionWithContext(ctx context.Context, opts *sql.TxOptions, callback func(tx Tx) error) error {
	return transaction(d.db, ctx, opts, callback)
}
