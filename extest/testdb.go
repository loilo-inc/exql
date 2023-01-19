package extest

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/loilo-inc/exql/v2"
)

func testDb(dsn string) exql.DB {
	db, err := exql.Open(&exql.OpenOptions{
		Url: dsn,
	})
	if err != nil {
		panic(err)
	}
	return db
}

var (
	DB    = TestDb()
	DB8   = TestDbMySQL8()
	SqlDB = TestSqlDB()
)

const (
	Dsn  = "root:@tcp(127.0.0.1:13326)/exql?charset=utf8mb4&parseTime=True&loc=Local"
	Dsn8 = "root:@tcp(127.0.0.1:13327)/exql?charset=utf8mb4&parseTime=True&loc=Local"
)

func TestDb() exql.DB {
	return testDb(Dsn)
}

func TestDbMySQL8() exql.DB {
	return testDb(Dsn8)
}

func TestSqlDB() *sql.DB {
	db, err := sql.Open("mysql", Dsn)
	if err != nil {
		panic(err)
	}
	return db
}
