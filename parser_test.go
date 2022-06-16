package exql

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser_ParseTable(t *testing.T) {
	t.Run("should return error when rows.Error() return error", func(t *testing.T) {
		mockDb, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDb.Close()

		p := &parser{}

		mock.ExpectQuery(`show columns from users`).WillReturnRows(
			sqlmock.NewRows([]string{"field", "type"}).
				AddRow("id", "int(11)").
				RowError(0, fmt.Errorf("err")))

		table, err := p.ParseTable(mockDb, "users")
		assert.Nil(t, table)
		assert.EqualError(t, err, "err")
	})
}

func TestParser_ParseType(t *testing.T) {
	p := &parser{}
	assertType := func(s string, nullable bool, tp interface{}) {
		ret, err := p.ParseType(s, nullable)
		assert.Nil(t, err)
		assert.Equal(t, ret, tp)
	}
	t.Run("int", func(t *testing.T) {
		list := [][]interface{}{
			{"int", "int64", "int64", "null.Int64", "null.Int64"},
			{"tinyint", "int64", "int64", "null.Int64", "null.Int64"},
			{"smallint", "int64", "int64", "null.Int64", "null.Int64"},
			{"mediumint", "int64", "int64", "null.Int64", "null.Int64"},
			{"bigint", "int64", "uint64", "null.Int64", "null.Uint64"},
		}
		for _, v := range list {
			title := v[0].(string)
			t.Run(title, func(t *testing.T) {
				assertType(fmt.Sprintf("%s(1)", title), false, v[1])
				assertType(fmt.Sprintf("%s(1) unsigned", title), false, v[2])
				assertType(fmt.Sprintf("%s(1)", title), true, v[3])
				assertType(fmt.Sprintf("%s(1) unsigned", title), true, v[4])
			})
		}
	})
	t.Run("float", func(t *testing.T) {
		assertType("float", false, "float32")
		assertType("float", true, "null.Float32")
	})
	t.Run("double", func(t *testing.T) {
		assertType("double", false, "float64")
		assertType("double", true, "null.Float64")
	})
	t.Run("date", func(t *testing.T) {
		list := [][]interface{}{
			{"date", "time.Time", "null.Time"},
			{"datetime", "time.Time", "null.Time"},
			{"datetime(6)", "time.Time", "null.Time"},
			{"timestamp", "time.Time", "null.Time"},
			{"timestamp(6)", "time.Time", "null.Time"},
			{"time", "string", "null.String"},
			{"time(6)", "string", "null.String"},
		}
		for _, v := range list {
			t := v[0].(string)
			assertType(t, false, v[1].(string))
			assertType(t, true, v[2].(string))
		}
	})
	t.Run("string", func(t *testing.T) {
		list := [][]interface{}{
			{"text", "string", "null.String"},
			{"tinytext", "string", "null.String"},
			{"mediumtext", "string", "null.String"},
			{"longtext", "string", "null.String"},
			{"char(10)", "string", "null.String"},
			{"varchar(255)", "string", "null.String"},
		}
		for _, v := range list {
			t := v[0].(string)
			assertType(t, false, v[1].(string))
			assertType(t, true, v[2].(string))
		}
	})
	t.Run("blob", func(t *testing.T) {
		list := [][]interface{}{
			{"blob", "[]byte", "null.Bytes"},
			{"tinyblob", "[]byte", "null.Bytes"},
			{"mediumblob", "[]byte", "null.Bytes"},
			{"longblob", "[]byte", "null.Bytes"},
		}
		for _, v := range list {
			t := v[0].(string)
			assertType(t, false, v[1].(string))
			assertType(t, true, v[2].(string))
		}
	})
	t.Run("json", func(t *testing.T) {
		assertType("json", false, "json.RawMessage")
		assertType("json", true, "null.JSON")
	})
}
