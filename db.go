package exql

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/loilo-inc/exql/v2/exdriver"
	q "github.com/loilo-inc/exql/v2/query"
)

type DB interface {
	Saver
	Mapper
	// Return *sql.DB instance
	DB() *sql.DB
	// Set db object
	SetDB(db *sql.DB)
	// Hooks returns hook manager for sql queries.
	// It panics if db connectin was not establised by exdriver.Connector,
	// because there's no way to hook queries without hacking driver.Connector.
	Hooks() *exdriver.HookList
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
	conn  *exdriver.Connector
	m     mapper
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
	// Use experimental hooks feature
	Experimental_Hooks bool
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
	var conn *exdriver.Connector
	var err error
	retryCnt := 0
	for retryCnt < maxRetryCount {
		d, conn, err = open(driverName, opts)
		if err != nil {
			log.Errorf("failed to connect database: %s, retrying after %ds...", err, int(retryInterval.Seconds()))
			<-time.NewTimer(retryInterval).C
			retryCnt++
		} else {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return &db{db: d, conn: conn, s: newSaver(d)}, nil
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
	return &db{db: d, s: newSaver(d)}
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
