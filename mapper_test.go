package exql

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/loilo-inc/exql/model"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null"
	"testing"
	"time"
)

func TestDb_MapRows(t *testing.T) {
	db := testDb()
	defer db.Close()
	t.Run("users", func(t *testing.T) {
		user1 := &model.Users{
			FirstName: null.StringFrom("user1"),
			LastName:  null.StringFrom("name"),
		}
		user2 := &model.Users{
			FirstName: null.StringFrom("user2"),
			LastName:  null.StringFrom("name"),
		}
		_, err := db.Insert(user1)
		_, err = db.Insert(user2)
		assert.Nil(t, err)
		defer func() {
			db.DB().Exec(`DELETE FROM users WHERE id = ?`, user1.Id)
			db.DB().Exec(`DELETE FROM users WHERE id = ?`, user2.Id)
		}()

		t.Run("struct", func(t *testing.T) {
			rows, err := db.DB().Query(`SELECT * FROM users LIMIT 1`)
			assert.Nil(t, err)
			defer rows.Close()
			var dest model.Users
			err = db.MapRows(rows, &dest)
			assert.Nil(t, err)
			assert.Equal(t, dest.FirstName.String, "user1")
			assert.Equal(t, dest.LastName.String, "name")
		})
		t.Run("slice", func(t *testing.T) {
			rows, err := db.DB().Query(`SELECT * FROM users LIMIT 2`)
			assert.Nil(t, err)
			defer rows.Close()
			var dest []*model.Users
			err = db.MapRows(rows, &dest)
			assert.Nil(t, err)
			assert.Equal(t, dest[0].FirstName.String, "user1")
			assert.Equal(t, dest[0].LastName.String, "name")
			assert.Equal(t, dest[1].FirstName.String, "user2")
			assert.Equal(t, dest[1].LastName.String, "name")
		})
	})
	t.Run("fields", func(t *testing.T) {
		now := time.Unix(time.Now().Unix(), 0)
		tinyBlob := []byte("tinyblob")
		mediumBlob := []byte("mediumblob")
		blob := []byte("blob")
		longblob := []byte("longblob")
		field := model.Fields{
			TinyintField:                   2,
			TinyintUnsignedField:           3,
			TinyintNullableField:           null.Int64From(4),
			TinyintUnsignedNullableField:   null.Int64From(5),
			SmallintField:                  6,
			SmallintUnsignedField:          7,
			SmallintNullableField:          null.Int64From(8),
			SmallintUnsignedNullableField:  null.Int64From(9),
			MediumintField:                 10,
			MediumintUnsignedField:         11,
			MediumintNullableField:         null.Int64From(12),
			MediumintUnsignedNullableField: null.Int64From(13),
			IntField:                       14,
			IntUnsignedField:               15,
			IntNullableField:               null.Int64From(16),
			IntUnsignedNullableField:       null.Int64From(17),
			BigintField:                    18,
			BigintUnsignedField:            19,
			BigintNullableField:            null.Int64From(20),
			BigintUnsignedNullableField:    null.Uint64From(21),
			FloatField:                     21,
			FloatNullField:                 null.Float32From(22),
			DoubleField:                    23,
			DoubleNullField:                null.Float64From(24),
			TinytextField:                  "tinytext",
			TinytextNullField:              null.StringFrom("tinytext"),
			MediumtextField:                "mediumtext",
			MediumtextNullField:            null.StringFrom("mediumtext"),
			TextField:                      "text",
			TextNullField:                  null.StringFrom("text"),
			LongtextField:                  "longtext",
			LongtextNullField:              null.StringFrom("longtext"),
			VarcharFiledField:              "varchar",
			VarcharNullField:               null.StringFrom("varchar"),
			CharFiledField:                 "char",
			CharFiledNullField:             null.StringFrom("char"),
			DateField:                      now,
			DateNullField:                  null.Time{},
			DatetimeField:                  now,
			DatetimeNullField:              null.Time{},
			TimeField:                      "12:34:56",
			TimeNullField:                  null.StringFrom("12:34:56"),
			TimestampField:                 now,
			TimestampNullField:             now,
			TinyblobField:                  tinyBlob,
			TinyblobNullField:              null.BytesFrom(tinyBlob),
			MediumblobField:                mediumBlob,
			MediumblobNullField:            null.BytesFrom(mediumBlob),
			BlobField:                      blob,
			BlobNullField:                  null.BytesFrom(blob),
			LongblobField:                  longblob,
			LongblobNullField:              null.BytesFrom(longblob),
		}
		_, err := db.Insert(&field)
		assert.False(t, field.Id == 0)
		assert.Nil(t, err)
		defer func() {
			db.DB().Exec(`DELETE FROM fields WHERE id = ?`, field.Id)
		}()
		t.Run("basic", func(t *testing.T) {
			rows, err := db.DB().Query(`SELECT * FROM fields WHERE id = ?`, field.Id)
			assert.Nil(t, err)
			var dest model.Fields
			err = db.MapRows(rows, &dest)
			assert.Nil(t, err)
			assert.Equal(t, dest.TinyintField, field.TinyintField)
			assert.Equal(t, dest.TinyintUnsignedField, field.TinyintUnsignedField)
			assert.Equal(t, dest.TinyintNullableField.Int64, field.TinyintNullableField.Int64)
			assert.Equal(t, dest.TinyintUnsignedNullableField.Int64, field.TinyintUnsignedNullableField.Int64)
			assert.Equal(t, dest.SmallintField, field.SmallintField)
			assert.Equal(t, dest.SmallintUnsignedField, field.SmallintUnsignedField)
			assert.Equal(t, dest.SmallintNullableField.Int64, field.SmallintNullableField.Int64)
			assert.Equal(t, dest.SmallintUnsignedNullableField.Int64, field.SmallintUnsignedNullableField.Int64)
			assert.Equal(t, dest.MediumintField, field.MediumintField)
			assert.Equal(t, dest.MediumintUnsignedField, field.MediumintUnsignedField)
			assert.Equal(t, dest.MediumintNullableField.Int64, field.MediumintNullableField.Int64)
			assert.Equal(t, dest.MediumintUnsignedNullableField.Int64, field.MediumintUnsignedNullableField.Int64)
			assert.Equal(t, dest.IntField, field.IntField)
			assert.Equal(t, dest.IntUnsignedField, field.IntUnsignedField)
			assert.Equal(t, dest.IntNullableField.Int64, field.IntNullableField.Int64)
			assert.Equal(t, dest.IntUnsignedNullableField.Int64, field.IntUnsignedNullableField.Int64)
			assert.Equal(t, dest.BigintField, field.BigintField)
			assert.Equal(t, dest.BigintUnsignedField, field.BigintUnsignedField)
			assert.Equal(t, dest.BigintNullableField.Int64, field.BigintNullableField.Int64)
			assert.Equal(t, dest.BigintUnsignedNullableField.Uint64, field.BigintUnsignedNullableField.Uint64)
			assert.Equal(t, dest.FloatField, field.FloatField)
			assert.Equal(t, dest.FloatNullField.Float32, field.FloatNullField.Float32)
			assert.Equal(t, dest.DoubleField, field.DoubleField)
			assert.Equal(t, dest.DoubleNullField.Float64, field.DoubleNullField.Float64)
			assert.Equal(t, dest.TinytextField, field.TinytextField)
			assert.Equal(t, dest.TinytextNullField.String, field.TinytextNullField.String)
			assert.Equal(t, dest.MediumtextField, field.MediumtextField)
			assert.Equal(t, dest.MediumtextNullField.String, field.MediumtextNullField.String)
			assert.Equal(t, dest.TextField, field.TextField)
			assert.Equal(t, dest.TextNullField.String, field.TextNullField.String)
			assert.Equal(t, dest.LongtextField, field.LongtextField)
			assert.Equal(t, dest.LongtextNullField.String, field.LongtextNullField.String)
			assert.Equal(t, dest.VarcharFiledField, field.VarcharFiledField)
			assert.Equal(t, dest.VarcharNullField.String, field.VarcharNullField.String)
			assert.Equal(t, dest.CharFiledField, field.CharFiledField)
			assert.Equal(t, dest.CharFiledNullField.String, field.CharFiledNullField.String)
			assert.Equal(t, dest.DateField.Format("2006-01-02"), field.DateField.Format("2006-01-02"))
			assert.Equal(t, dest.DateNullField.Time.Unix(), field.DateNullField.Time.Unix())
			assert.Equal(t, dest.DatetimeField.Unix(), field.DatetimeField.Unix())
			assert.Equal(t, dest.DatetimeNullField.Time.Unix(), field.DatetimeNullField.Time.Unix())
			assert.Equal(t, dest.TimeField, field.TimeField)
			assert.Equal(t, dest.TimeNullField.String, field.TimeNullField.String)
			assert.Equal(t, dest.TimestampField.Unix(), field.TimestampField.Unix())
			assert.Equal(t, dest.TimestampNullField.Unix(), field.TimestampNullField.Unix())
			assert.ElementsMatch(t, dest.TinyblobField, field.TinyblobField)
			assert.ElementsMatch(t, dest.TinyblobNullField.Bytes, field.TinyblobNullField.Bytes)
			assert.ElementsMatch(t, dest.MediumblobField, field.MediumblobField)
			assert.ElementsMatch(t, dest.MediumblobNullField.Bytes, field.MediumblobNullField.Bytes)
			assert.ElementsMatch(t, dest.BlobField, field.BlobField)
			assert.ElementsMatch(t, dest.BlobNullField.Bytes, field.BlobNullField.Bytes)
			assert.ElementsMatch(t, dest.LongblobField, field.LongblobField)
			assert.ElementsMatch(t, dest.LongblobNullField.Bytes, field.LongblobNullField.Bytes)
		})
	})
}
