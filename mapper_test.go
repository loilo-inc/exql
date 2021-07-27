package exql

import (
	"encoding/json"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/loilo-inc/exql/model"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null"
)

type partialUser struct {
	Id       int64       `exql:"column:id;primary"`
	LastName null.String `exql:"column:last_name"`
}

func setupUsers(t *testing.T, db DB) ([]*model.Users, func()) {
	user1 := &model.Users{
		FirstName: null.StringFrom("user1"),
		LastName:  null.StringFrom("name"),
	}
	user2 := &model.Users{
		FirstName: null.StringFrom("user2"),
		LastName:  null.StringFrom("name"),
	}
	_, err := db.Insert(user1)
	assert.Nil(t, err)
	_, err = db.Insert(user2)
	assert.Nil(t, err)
	return []*model.Users{user1, user2}, func() {
		db.DB().Exec(`DELETE FROM users WHERE id = ?`, user1.Id)
		db.DB().Exec(`DELETE FROM users WHERE id = ?`, user2.Id)
	}
}
func setupFields(t *testing.T, db DB) (*model.Fields, func()) {
	now := time.Unix(time.Now().Unix(), 0)
	tinyBlob := []byte("tinyblob")
	mediumBlob := []byte("mediumblob")
	blob := []byte("blob")
	longblob := []byte("longblob")
	rawJson := json.RawMessage(`{"string":"json","number":123.456,"boolean":true,"array":["Apple","Orange"]}`)
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
		TimestampNullField:             null.Time{},
		TinyblobField:                  tinyBlob,
		TinyblobNullField:              null.BytesFrom(tinyBlob),
		MediumblobField:                mediumBlob,
		MediumblobNullField:            null.BytesFrom(mediumBlob),
		BlobField:                      blob,
		BlobNullField:                  null.BytesFrom(blob),
		LongblobField:                  longblob,
		LongblobNullField:              null.BytesFrom(longblob),
		JsonField:                      rawJson,
		JsonNullField:                  null.JSONFrom(rawJson),
	}
	_, err := db.Insert(&field)
	assert.False(t, field.Id == 0)
	assert.Nil(t, err)
	return &field, func() {
		db.DB().Exec(`DELETE FROM fields WHERE id = ?`, field.Id)
	}
}
func assertFields(t *testing.T, dest *model.Fields, field *model.Fields) {
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
	assert.Equal(t, dest.TimestampNullField.Time.Unix(), field.TimestampNullField.Time.Unix())
	assert.ElementsMatch(t, dest.TinyblobField, field.TinyblobField)
	assert.ElementsMatch(t, dest.TinyblobNullField.Bytes, field.TinyblobNullField.Bytes)
	assert.ElementsMatch(t, dest.MediumblobField, field.MediumblobField)
	assert.ElementsMatch(t, dest.MediumblobNullField.Bytes, field.MediumblobNullField.Bytes)
	assert.ElementsMatch(t, dest.BlobField, field.BlobField)
	assert.ElementsMatch(t, dest.BlobNullField.Bytes, field.BlobNullField.Bytes)
	assert.ElementsMatch(t, dest.LongblobField, field.LongblobField)
	assert.ElementsMatch(t, dest.LongblobNullField.Bytes, field.LongblobNullField.Bytes)
	assert.JSONEq(t, string(dest.JsonField), string(field.JsonField))
	assert.JSONEq(t, string(dest.JsonNullField.JSON), string(field.JsonNullField.JSON))
}
func TestMapper_MapMany(t *testing.T) {
	db := testDb()
	defer db.Close()
	m := &mapper{}
	t.Run("users", func(t *testing.T) {
		users, reset := setupUsers(t, db)
		defer reset()
		t.Run("basic", func(t *testing.T) {
			rows, err := db.DB().Query(`SELECT * FROM users WHERE id in (?,?) ORDER BY id`, users[0].Id, users[1].Id)
			assert.Nil(t, err)
			defer rows.Close()
			var dest []*model.Users
			err = m.MapMany(rows, &dest)
			assert.Nil(t, err)
			assert.Equal(t, dest[0].FirstName.String, users[0].FirstName.String)
			assert.Equal(t, dest[0].LastName.String, users[0].LastName.String)
			assert.Equal(t, dest[1].FirstName.String, users[1].FirstName.String)
			assert.Equal(t, dest[1].LastName.String, users[1].LastName.String)
		})
	})
	t.Run("fields", func(t *testing.T) {
		field, reset := setupFields(t, db)
		defer reset()
		t.Run("basic", func(t *testing.T) {
			rows, err := db.DB().Query(`SELECT * FROM fields WHERE id = ?`, field.Id)
			assert.Nil(t, err)
			var dest []*model.Fields
			err = m.MapMany(rows, &dest)
			assert.Nil(t, err)
			assertFields(t, dest[0], field)
		})
	})
	t.Run("should return error if destination is not pointer of slice of pointer of struct", func(t *testing.T) {
		doTest := func(i interface{}) {
			assert.EqualError(t, m.MapMany(nil, i),
				"destination must be pointer of slice of struct",
			)
		}
		t.Run("int", func(t *testing.T) {
			doTest(0)
		})
		t.Run("*int", func(t *testing.T) {
			i := 0
			doTest(&i)
		})
		t.Run("[]struct", func(t *testing.T) {
			var i []model.Users
			doTest(&i)
		})
		t.Run("struct", func(t *testing.T) {
			var i model.Users
			doTest(i)
		})
		t.Run("*struct", func(t *testing.T) {
			var i model.Users
			doTest(&i)
		})
		t.Run("nil", func(t *testing.T) {
			doTest(nil)
		})
	})
	t.Run("should return exql.ErrRecordNotFound if rows is empty", func(t *testing.T) {
		rows, err := db.DB().Query(`SELECT * FROM users where id = -1`)
		assert.Nil(t, err)
		var dest []*model.Users
		err = m.MapMany(rows, &dest)
		assert.Equal(t, ErrRecordNotFound, err)
	})
}

func TestMapper_Map(t *testing.T) {
	db := testDb()
	m := &mapper{}
	t.Run("users", func(t *testing.T) {
		users, reset := setupUsers(t, db)
		defer reset()
		t.Run("basic", func(t *testing.T) {
			rows, err := db.DB().Query(
				`SELECT * FROM users WHERE id IN (?, ?) ORDER BY id`,
				users[0].Id, users[1].Id,
			)
			assert.Nil(t, err)
			defer rows.Close()
			var dest model.Users
			err = m.Map(rows, &dest)
			assert.Nil(t, err)
			assert.Equal(t, dest.FirstName.String, users[0].FirstName.String)
			assert.Equal(t, dest.LastName.String, users[0].LastName.String)
		})
		t.Run("partial", func(t *testing.T) {
			user := users[0]
			rows, err := db.DB().Query("SELECT * FROM users WHERE id = ?", user.Id)
			assert.Nil(t, err)
			var p partialUser
			err = m.Map(rows, &p)
			assert.Nil(t, err)
			assert.Equal(t, user.Id, p.Id)
			assert.Equal(t, user.LastName.String, p.LastName.String)
		})
	})
	t.Run("fields", func(t *testing.T) {
		field, reset := setupFields(t, db)
		defer reset()
		t.Run("basic", func(t *testing.T) {
			rows, err := db.DB().Query("SELECT * FROM fields WHERE id = ?", field.Id)
			assert.Nil(t, err)
			var dest model.Fields
			err = m.Map(rows, &dest)
			assert.Nil(t, err)
			assertFields(t, &dest, field)
		})
	})
	t.Run("should return error if destination is not pointer of struct", func(t *testing.T) {
		doTest := func(i interface{}) {
			assert.EqualError(t, m.Map(nil, i), "destination must be pointer of struct")
		}
		t.Run("int", func(t *testing.T) {
			doTest(0)
		})
		t.Run("*int", func(t *testing.T) {
			i := 0
			doTest(&i)
		})
		t.Run("slice", func(t *testing.T) {
			var i []*model.Users
			doTest(&i)
		})
		t.Run("*slice", func(t *testing.T) {
			var i []*model.Users
			doTest(&i)
		})
		t.Run("nil", func(t *testing.T) {
			doTest(nil)
		})
	})

	t.Run("should return exql.ErrRecordNotFound if rows is empty", func(t *testing.T) {
		rows, err := db.DB().Query(`SELECT * FROM users where id = -1`)
		assert.Nil(t, err)
		var dest model.Users
		err = m.Map(rows, &dest)
		assert.Equal(t, ErrRecordNotFound, err)
	})
}

func TestDb_MapRowsSerial(t *testing.T) {
	db := testDb()
	defer db.Close()

	user1 := &model.Users{
		FirstName: null.StringFrom("user1"),
		LastName:  null.StringFrom("name"),
	}
	user2 := &model.Users{
		FirstName: null.StringFrom("user2"),
		LastName:  null.StringFrom("name"),
	}
	user3 := &model.Users{
		FirstName: null.StringFrom("user3"),
		LastName:  null.StringFrom("name"),
	}
	_, err := db.Insert(user1)
	assert.Nil(t, err)
	_, err = db.Insert(user2)
	assert.Nil(t, err)
	_, err = db.Insert(user3)
	assert.Nil(t, err)
	group := &model.UserGroups{
		Name: "group1",
	}
	_, err = db.Insert(group)
	assert.Nil(t, err)
	member1 := &model.GroupUsers{
		UserId:  user1.Id,
		GroupId: group.Id,
	}
	member2 := &model.GroupUsers{
		UserId:  user2.Id,
		GroupId: group.Id,
	}
	_, err = db.Insert(member1)
	assert.Nil(t, err)
	_, err = db.Insert(member2)
	assert.Nil(t, err)
	defer func() {
		db.DB().Exec(`DELETE FROM users WHERE id IN (?,?,?)`, user1.Id, user2.Id, user3.Id)
		db.DB().Exec(`DELETE FROM user_groups WHERE id = ?`, group.Id)
		db.DB().Exec(`DELETE from group_users WHERE id IN (?,?)`, member1.Id, member1.Id)
	}()
	m := NewSerialMapper(func(i int) string {
		return "id"
	})
	t.Run("basic", func(t *testing.T) {
		query := `
SELECT * FROM users 
JOIN group_users on group_users.user_id = users.id
JOIN user_groups on group_users.group_id = user_groups.id
WHERE user_groups.id = ?
`
		rows, err := db.DB().Query(query, group.Id)
		assert.Nil(t, err)
		var users []*model.Users
		for rows.Next() {
			var group model.UserGroups
			var user model.Users
			var mem model.GroupUsers
			err := m.Map(rows, &user, &group, &mem)
			assert.Nil(t, err)
			users = append(users, &user)
		}
		assert.Nil(t, err)
		assert.Equal(t, 2, len(users))
		if users == nil {
			t.Fail()
			return
		}
		assert.Equal(t, user1.Id, users[0].Id)
		assert.Equal(t, user2.Id, users[1].Id)
	})
	t.Run("subset", func(t *testing.T) {
		query := `
SELECT users.*, user_groups.* FROM users 
JOIN group_users on group_users.user_id = users.id
JOIN user_groups on group_users.group_id = user_groups.id
WHERE user_groups.id = ?
`
		rows, err := db.DB().Query(query, group.Id)
		assert.Nil(t, err)
		var users []*model.Users
		var groups []*model.UserGroups
		for rows.Next() {
			var group model.UserGroups
			var user model.Users
			err := m.Map(rows, &user, &group)
			assert.Nil(t, err)
			users = append(users, &user)
			groups = append(groups, &group)
		}
		assert.Nil(t, err)
		assert.Equal(t, 2, len(users))
		if users == nil || groups == nil {
			t.Fail()
			return
		}
		assert.Equal(t, user1.Id, users[0].Id)
		assert.Equal(t, user2.Id, users[1].Id)
		assert.Equal(t, group.Id, groups[0].Id)
	})
	t.Run("outer join", func(t *testing.T) {
		query := `
SELECT users.*, user_groups.* FROM users
LEFT JOIN group_users on group_users.user_id = users.id
LEFT JOIN user_groups on group_users.group_id = user_groups.id
WHERE users.id IN (?, ?)
ORDER BY users.id
`
		rows, err := db.DB().Query(query, user1.Id, user3.Id)

		assert.Nil(t, err)
		var users []*model.Users
		var groups []*model.UserGroups
		for rows.Next() {
			var group *model.UserGroups
			var user *model.Users
			err := m.Map(rows, &user, &group)
			assert.Nil(t, err)
			users = append(users, user)
			groups = append(groups, group)
		}
		assert.Nil(t, err)
		assert.Equal(t, 2, len(users))
		assert.Equal(t, 2, len(groups))
		assert.Equal(t, user1, users[0])
		assert.Equal(t, group, groups[0])
		assert.Equal(t, user3, users[1])
		assert.Equal(t, (*model.UserGroups)(nil), groups[1])
	})

	t.Run("partial", func(t *testing.T) {
		query := `
SELECT users.*, user_groups.* FROM users
LEFT JOIN group_users on group_users.user_id = users.id
LEFT JOIN user_groups on group_users.group_id = user_groups.id
WHERE users.id IN (?, ?)
ORDER BY users.id
`
		rows, err := db.DB().Query(query, user1.Id, user3.Id)

		assert.Nil(t, err)
		var users []*partialUser
		var groups []*model.UserGroups
		for rows.Next() {
			var group *model.UserGroups
			var user *partialUser
			err := m.Map(rows, &user, &group)
			assert.Nil(t, err)
			users = append(users, user)
			groups = append(groups, group)
		}
		assert.Nil(t, err)
		assert.Equal(t, 2, len(users))
		assert.Equal(t, 2, len(groups))
		assert.Equal(t, user1.Id, users[0].Id)
		assert.Equal(t, group, groups[0])
		assert.Equal(t, user3.Id, users[1].Id)
		assert.Equal(t, (*model.UserGroups)(nil), groups[1])
	})

	t.Run("should return error if head column is not found", func(t *testing.T) {
		t.Run("inner join case", func(t *testing.T) {
			query := `
SELECT users.*, user_groups.* FROM users 
JOIN group_users on group_users.user_id = users.id
JOIN user_groups on group_users.group_id = user_groups.id
WHERE user_groups.id = ? ORDER BY users.id LIMIT 1
`
			rows, err := db.DB().Query(query, group.Id)
			assert.Nil(t, err)
			m := &serialMapper{splitter: func(i int) string {
				return "var"
			}}
			for rows.Next() {
				var user model.Users
				var ug model.UserGroups
				err := m.Map(rows, &user, &ug)
				assert.EqualError(t, err, "head col mismatch: expected=var, actual=id")
				break
			}
		})
		t.Run("outer join case", func(t *testing.T) {
			query := `
SELECT users.*, user_groups.* FROM users 
LEFT JOIN group_users on group_users.user_id = users.id
LEFT JOIN user_groups on group_users.group_id = user_groups.id
WHERE user_groups.id = ? ORDER BY users.id LIMIT 1
`
			rows, err := db.DB().Query(query, group.Id)
			assert.Nil(t, err)
			m := &serialMapper{splitter: func(i int) string {
				return "var"
			}}
			for rows.Next() {
				var user *model.Users
				var ug *model.UserGroups
				err := m.Map(rows, &user, &ug)
				assert.EqualError(t, err, "head col mismatch: expected=var, actual=id")
				break
			}

		})
	})
	t.Run("should return error if dest is empty", func(t *testing.T) {
		err := m.Map(nil)
		assert.EqualError(t, err, "empty dest list")
	})
	t.Run("should return error if destination is invalid", func(t *testing.T) {
		doTest := func(i ...interface{}) {
			assert.Equal(t, ErrMapRowSerialDestination, m.Map(nil, i...))
		}
		t.Run("int", func(t *testing.T) {
			doTest(0, 1, 2)
		})
		t.Run("*int", func(t *testing.T) {
			i := 0
			doTest(&i, &i)
		})
		t.Run("slice", func(t *testing.T) {
			var i []*model.Users
			doTest(&i, &i)
		})
		t.Run("*slice", func(t *testing.T) {
			var i []*model.Users
			doTest(&i, &i)
		})
		t.Run("nil", func(t *testing.T) {
			doTest(nil, nil)
		})
		t.Run("***struct", func(t *testing.T) {
			var user **model.Users
			var group **model.GroupUsers
			doTest(&user, &group)
		})
		t.Run("non nil **struct", func(t *testing.T) {
			var user = &model.Users{}
			doTest(&user)
		})
	})
}
