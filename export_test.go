package exql

import (
	"database/sql"
	"fmt"
	"reflect"

	_ "github.com/go-sql-driver/mysql"
)

const dbUrl = "root:@tcp(127.0.0.1:13326)/exql?charset=utf8mb4&parseTime=True&loc=Local"

func testDb() DB {
	db, err := Open(&OpenOptions{
		Url: dbUrl,
	})
	if err != nil {
		panic(err)
	}
	return db
}

func testSqlDB() *sql.DB {
	db, err := sql.Open("mysql", dbUrl)
	if err != nil {
		panic(err)
	}
	return db
}

type errReflector struct {
}

var _ Reflector = (*errReflector)(nil)

func (r *errReflector) GetSchema(modelPtr any) (*modelSchema, error) {
	return nil, fmt.Errorf("error reflector")
}

func (r *errReflector) GetSchemaFromValue(destValue *reflect.Value, _ bool) (*modelSchema, error) {
	return nil, fmt.Errorf("error reflector")
}

func (r *errReflector) ClearSchemaCache() {}
