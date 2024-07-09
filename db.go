package exql

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"log"

	"golang.org/x/xerrors"
)

type DB interface {
	Saver
	Mapper
	Finder
	// DB returns *sql.DB object.
	DB() *sql.DB
	// SetDB sets *sql.DB object.
	SetDB(db *sql.DB)
	// Transaction begins a transaction and commits after the callback is called.
	// If an error is returned from the callback, it is rolled back.
	// Internally call tx.BeginTx(context.Background(), nil).
	Transaction(callback func(tx Tx) error) error
	// TransactionWithContext is same as Transaction().
	// Internally call tx.BeginTx(ctx, opts).
	TransactionWithContext(ctx context.Context, opts *sql.TxOptions, callback func(tx Tx) error) error
	// Close calls db.Close().
	Close() error
}

type db struct {
	*saver
	*finder
	*mapper
	db    *sql.DB
	mutex sync.Mutex
}

// OpenFunc is an abstraction of sql.Open function.
type OpenFunc func(driverName string, url string) (*sql.DB, error)

type OpenOptions struct {
	// @required
	// DSN format for database connection.
	Url string
	// @default "mysql"
	DriverName string
	// @default 5
	MaxRetryCount int
	// @default 5s
	RetryInterval time.Duration
	// Custom opener function.
	OpenFunc OpenFunc
}

// Open opens the connection to the database and makes exql.DB interface.
func Open(opts *OpenOptions) (DB, error) {
	return OpenContext(context.Background(), opts)
}

// OpenContext opens the connection to the database and makes exql.DB interface.
// If something failed, it retries automatically until given retry strategies satisfied
// or aborts handshaking.
//
// Example:
//
//	db, err := exql.Open(context.Background(), &exql.OpenOptions{
//		Url: "user:pass@tcp(127.0.0.1:3306)/database?charset=utf8mb4&parseTime=True&loc=Local",
//		MaxRetryCount: 3,
//		RetryInterval: 10, //sec
//	})
func OpenContext(ctx context.Context, opts *OpenOptions) (DB, error) {
	if opts.Url == "" {
		return nil, xerrors.New("opts.Url is required")
	}
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
	var openFunc OpenFunc = sql.Open
	if opts.OpenFunc != nil {
		openFunc = opts.OpenFunc
	}
	retryCnt := 0
	for retryCnt < maxRetryCount {
		d, err = openFunc(driverName, opts.Url)
		if err != nil {
			goto retry
		} else if err = d.PingContext(ctx); err != nil {
			goto retry
		} else {
			goto success
		}
	retry:
		log.Printf("failed to connect database: %s, retrying after %ds...\n", err, int(retryInterval.Seconds()))
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
		saver:  newSaver(d),
		finder: newFinder(d),
		mapper: &mapper{},
		db:     d,
	}
}

func (d *db) Close() error {
	return d.db.Close()
}

func (d *db) DB() *sql.DB {
	return d.db
}

func (d *db) SetDB(db *sql.DB) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.db = db
	d.saver.ex = db
}

func (d *db) Transaction(callback func(tx Tx) error) error {
	return d.TransactionWithContext(context.Background(), nil, callback)
}

func (d *db) TransactionWithContext(ctx context.Context, opts *sql.TxOptions, callback func(tx Tx) error) error {
	return Transaction(d.db, ctx, opts, callback)
}
