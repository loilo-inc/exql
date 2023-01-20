package exql

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/apex/log"
	q "github.com/loilo-inc/exql/v2/query"
)

type DB interface {
	Saver
	Mapper
	// DB returns *sql.DB object.
	DB() *sql.DB
	// SetDB sets *sql.DB object.
	SetDB(db *sql.DB)
	// Transaction begins transaction and commits after callback called.
	// If error returned from callback, transaction is rolled back.
	// Internally call tx.BeginTx(context.Background(), nil).
	Transaction(callback func(tx Tx) error) error
	// TransactionWithContext is same as Transaction().
	// Internally call tx.BeginTx(ctx, opts).
	TransactionWithContext(ctx context.Context, opts *sql.TxOptions, callback func(tx Tx) error) error
	// Close calls db.Close().
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
	// DSN format for database connection.
	Url string
	// @default 5
	MaxRetryCount int
	// @default 5s
	RetryInterval time.Duration
}

// Open opens connection to database and makes exql.DB interface.
// If something failed, it retries automatically until given retry strategies satisfied
// or aborts handshaking.
//
// Example:
//
//	db, err := exql.Open(&exql.OpenOptions{
//		Url: "user:pass@tcp(127.0.0.1:3306)/database?charset=utf8mb4&parseTime=True&loc=Local",
//		MaxRetryCount: 3,
//		RetryInterval: 10, //sec
//	})
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

func (d *db) Insert(modelPtr Model) (sql.Result, error) {
	return d.s.Insert(modelPtr)
}

func (d *db) InsertContext(ctx context.Context, modelPtr Model) (sql.Result, error) {
	return d.s.InsertContext(ctx, modelPtr)
}

func (d *db) Update(table string, set map[string]interface{}, where q.Condition) (sql.Result, error) {
	return d.s.Update(table, set, where)
}

func (d *db) UpdateModel(ptr ModelUpdate, where q.Condition) (sql.Result, error) {
	return d.s.UpdateModel(ptr, where)
}

func (d *db) UpdateContext(ctx context.Context, table string, set map[string]interface{}, where q.Condition) (sql.Result, error) {
	return d.s.UpdateContext(ctx, table, set, where)
}

func (d *db) UpdateModelContext(ctx context.Context, ptr ModelUpdate, where q.Condition) (sql.Result, error) {
	return d.s.UpdateModelContext(ctx, ptr, where)
}

func (d *db) Delete(table string, where q.Condition) (sql.Result, error) {
	return d.s.Delete(table, where)
}

func (d *db) DeleteContext(ctx context.Context, table string, where q.Condition) (sql.Result, error) {
	return d.s.DeleteContext(ctx, table, where)
}

func (d *db) Exec(query q.Query) (sql.Result, error) {
	return d.s.Exec(query)
}

func (d *db) ExecContext(ctx context.Context, query q.Query) (sql.Result, error) {
	return d.s.ExecContext(ctx, query)
}

func (d *db) Query(query q.Query) (*sql.Rows, error) {
	return d.s.Query(query)
}

func (d *db) QueryContext(ctx context.Context, query q.Query) (*sql.Rows, error) {
	return d.s.QueryContext(ctx, query)
}

func (d *db) QueryRow(query q.Query) (*sql.Row, error) {
	return d.s.QueryRow(query)
}

func (d *db) QueryRowContext(ctx context.Context, query q.Query) (*sql.Row, error) {
	return d.s.QueryRowContext(ctx, query)
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
	return Transaction(d.db, ctx, opts, callback)
}
