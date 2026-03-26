package benchmark

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

const dbUrl = "root:@tcp(127.0.0.1:13326)/exql?charset=utf8mb4&parseTime=True&loc=Local"

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

func testSqlDB(t testing.TB) *sql.DB {
	db, err := sql.Open("mysql", dbUrl)
	if err != nil {
		t.Fatal(err)
	}
	return db
}
