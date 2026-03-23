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

func resetTestDB(db *sql.DB) error {
	db.Exec("SET FOREIGN_KEY_CHECKS=0")
	defer db.Exec("SET FOREIGN_KEY_CHECKS=1")

	_, err := db.Exec("TRUNCATE TABLE `group_users`")
	if err != nil {
		return err
	}
	_, err = db.Exec("TRUNCATE TABLE `users`")
	if err != nil {
		return err
	}
	_, err = db.Exec("TRUNCATE TABLE `user_groups`")
	if err != nil {
		return err
	}
	_, err = db.Exec("TRUNCATE TABLE `fields`")
	if err != nil {
		return err
	}
	return nil
}

type errReflector struct {
}

var _ Reflector = (*errReflector)(nil)

func (r *errReflector) GetSchema(modelPtr reflect.Type, forUpdate bool) (*modelSchema, error) {
	return nil, fmt.Errorf("error reflector")
}

func (r *errReflector) GetModelSchema(dest any, forUpdate bool) (*modelSchema, error) {
	return nil, fmt.Errorf("error reflector")
}

func (r *errReflector) ClearSchemaCache() {}
