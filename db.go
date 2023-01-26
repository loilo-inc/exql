package exql

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"log"

	"github.com/loilo-inc/exql/v2/exdriver"

	q "github.com/loilo-inc/exql/v2/query"
)

type DB interface {
	Saver
	Mapper
	Finder
	// DB returns *sql.DB object.
	DB() *sql.DB
	// SetDB sets *sql.DB object.
	SetDB(db *sql.DB)
	// Hooks returns hook manager for sql queries.
	// It panics if db connectin was not establised by exdriver.Connector,
	// because there's no way to hook queries without hacking driver.Connector.
	Hooks() *exdriver.HookList
	// Internally call tx.BeginTx(context.Background(), nil)
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
	db    *sql.DB
	s     *saver
	conn  *exdriver.Connector
	f     *finder
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
	// Use experimental hooks feature
	Experimental_Hooks bool
}

// Open opens the connection to the database and makes exql.DB interface.
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
	var conn *exdriver.Connector
	var err error
	retryCnt := 0
	for retryCnt < maxRetryCount {
		d, conn, err = open(driverName, opts)
		if err != nil {
			log.Printf("failed to connect database: %s, retrying after %ds...\n", err, int(retryInterval.Seconds()))
			<-time.NewTimer(retryInterval).C
			retryCnt++
		} else {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return newDB(d, conn), nil
}

func open(driverName string, opts *OpenOptions) (*sql.DB, *exdriver.Connector, error) {
	d, err := sql.Open(driverName, opts.Url)
	if err != nil {
		return nil, nil, err
	}
	var conn *exdriver.Connector
	if opts.Experimental_Hooks {
		conn = exdriver.NewConnector(d.Driver(), opts.Url)
		d = sql.OpenDB(conn)
	}
	if err = d.Ping(); err != nil {
		return nil, nil, err
	}
	return d, conn, nil
}

func NewDB(d *sql.DB) DB {
	return &db{
		db: d,
		s:  &saver{ex: d},
		f:  newFinder(d),
	}
}
func newDB(d *sql.DB, conn *exdriver.Connector) *db {
	return &db{db: d, conn: conn, f: newFinder(d), s: newSaver(d)}
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

func (d *db) Hooks() *exdriver.HookList {
	if d.conn == nil {
		panic("hooks is disabled because there's no hooked connector")
	}
	return d.conn.Hooks()
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
	d.conn = nil
	d.s = newSaver(db)
}

func (d *db) Transaction(callback func(tx Tx) error) error {
	return d.TransactionWithContext(context.Background(), nil, callback)
}

func (d *db) TransactionWithContext(ctx context.Context, opts *sql.TxOptions, callback func(tx Tx) error) error {
	return Transaction(d.db, ctx, opts, callback)
}

// Find implements DB
func (d *db) Find(q q.Query, destPtrOfStruct any) error {
	return d.f.Find(q, destPtrOfStruct)
}

// FindContext implements DB
func (d *db) FindContext(ctx context.Context, q q.Query, destPtrOfStruct any) error {
	return d.f.FindContext(ctx, q, destPtrOfStruct)
}

// FindMany implements DB
func (d *db) FindMany(q q.Query, destSlicePtrOfStruct any) error {
	return d.f.FindMany(q, destSlicePtrOfStruct)
}

// FindManyContext implements DB
func (d *db) FindManyContext(ctx context.Context, q q.Query, destSlicePtrOfStruct any) error {
	return d.f.FindManyContext(ctx, q, destSlicePtrOfStruct)
}

// Deprecated: Use Find or MapRow. It will be removed in next version.
func (d *db) Map(rows *sql.Rows, destPtr any) error {
	return MapRow(rows, destPtr)
}

// Deprecated: Use FindContext or MapRows. It will be removed in next version.
func (d *db) MapMany(rows *sql.Rows, destSlicePtr any) error {
	return MapRows(rows, destSlicePtr)
}
