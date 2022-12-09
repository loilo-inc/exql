//go:generate mockgen -source $GOFILE -destination ./mocks/mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE
package exql

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/apex/log"
	q "github.com/loilo-inc/exql/query"
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

// An abstraction of sql.DB/sql.Tx
type Executor interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type db struct {
	Db     *sql.DB
	Saver  *saver
	Mapper *mapper
	mutex  sync.Mutex
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
		Db:     d,
		Saver:  &saver{ex: d},
		Mapper: &mapper{},
	}
}

func (d *db) Insert(structPtr interface{}) (sql.Result, error) {
	return d.Saver.Insert(structPtr)
}

func (d *db) InsertContext(ctx context.Context, structPtr interface{}) (sql.Result, error) {
	return d.Saver.InsertContext(ctx, structPtr)
}

func (d *db) Update(table string, set map[string]interface{}, where q.Stmt) (sql.Result, error) {
	return d.Saver.Update(table, set, where)
}

func (d *db) UpdateModel(ptr interface{}, where q.Stmt) (sql.Result, error) {
	return d.Saver.UpdateModel(ptr, where)
}

func (d *db) UpdateContext(ctx context.Context, table string, set map[string]interface{}, where q.Stmt) (sql.Result, error) {
	return d.Saver.UpdateContext(ctx, table, set, where)
}

func (d *db) UpdateModelContext(ctx context.Context, ptr interface{}, where q.Stmt) (sql.Result, error) {
	return d.Saver.UpdateModelContext(ctx, ptr, where)
}

func (d *db) Delete(table string, where q.Stmt) (sql.Result, error) {
	return d.Saver.Delete(table, where)
}

func (d *db) DeleteContext(ctx context.Context, table string, where q.Stmt) (sql.Result, error) {
	return d.Saver.DeleteContext(ctx, table, where)
}

func (d *db) Map(rows *sql.Rows, pointerOfStruct interface{}) error {
	return d.Mapper.Map(rows, pointerOfStruct)
}

func (d *db) MapMany(rows *sql.Rows, pointerOfSliceOfStruct interface{}) error {
	return d.Mapper.MapMany(rows, pointerOfSliceOfStruct)
}

func (d *db) Close() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.Db.Close()
}

func (d *db) DB() *sql.DB {
	return d.Db
}

func (d *db) SetDB(db *sql.DB) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.Db = db
	d.Saver.ex = db
}

func (d *db) Transaction(callback func(tx Tx) error) error {
	return d.TransactionWithContext(context.Background(), nil, callback)
}

func (d *db) TransactionWithContext(ctx context.Context, opts *sql.TxOptions, callback func(tx Tx) error) error {
	return Transaction(d.Db, ctx, opts, callback)
}
