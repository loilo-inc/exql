package exql

import (
	"testing"

	v2 "github.com/loilo-inc/exql/v2"
	v2model "github.com/loilo-inc/exql/v2/model"
	"github.com/loilo-inc/exql/v3/internal/mock"
	"github.com/loilo-inc/exql/v3/model"
	"github.com/loilo-inc/exql/v3/query"
)

func BenchmarkQueryForInsert(b *testing.B) {
	v2user := v2model.Users{}
	v3user := model.Users{}
	refl := newReflector()
	b.Run("v2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v2.QueryForInsert(&v2user)
		}
	})
	b.Run("v3", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			queryForInsert(refl, &v3user)
		}
	})
}

func BenchmarkQueryForBulkInsert(b *testing.B) {
	var v2users []*v2model.Users
	var v3users []*model.Users
	for range 100 {
		v2users = append(v2users, &v2model.Users{})
		v3users = append(v3users, &model.Users{})
	}
	refl := newReflector()
	b.Run("v2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v2.QueryForBulkInsert(v2users...)
		}
	})
	b.Run("v3", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			queryForBulkInsert(refl, v3users...)
		}
	})
}

func BenchmarkQueryForUpdate(b *testing.B) {
	v2user := v2model.UpdateUsers{}
	v3user := model.UpdateUsers{}
	v2where := v2.Where("id = ?", 1)
	v3where := Where("id = ?", 1)
	refl := newReflector()
	b.Run("v2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v2.QueryForUpdateModel(&v2user, v2where)
		}
	})
	b.Run("v3", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			queryForUpdateModel(refl, &v3user, v3where)
		}
	})
}

func BenchmarkMapRow(b *testing.B) {
	refl := newReflector()
	noRelf := noCacheReflector
	b.Run("no-cache", func(b *testing.B) {
		row := &mock.Rows{
			Cols:   []string{"id", "name", "age"},
			Values: [][]any{{1, "exql", 6}},
		}
		for range b.N {
			var user model.Users
			row.Idx = 0
			mapRow(noRelf, row, &user)
		}
	})
	b.Run("cache", func(b *testing.B) {
		row := &mock.Rows{
			Cols:   []string{"id", "name", "age"},
			Values: [][]any{{1, "exql", 6}},
		}
		for range b.N {
			var user model.Users
			row.Idx = 0
			mapRow(refl, row, &user)
		}
	})
}

func BenchmarkMapRows(b *testing.B) {
	refl := newReflector()
	noRelf := noCacheReflector
	rows := &mock.Rows{
		Cols: []string{"id", "name", "age"},
	}
	for range 100 {
		rows.Values = append(rows.Values, []any{1, "exql", 6})
	}
	b.Run("no-cache", func(b *testing.B) {
		for range b.N {
			var users []*model.Users
			rows.Idx = 0
			mapRows(noRelf, rows, &users)
		}
	})
	b.Run("cache", func(b *testing.B) {
		for range b.N {
			var users []*model.Users
			rows.Idx = 0
			mapRows(refl, rows, &users)
		}
	})
}

func BenchmarkInsert(b *testing.B) {
	sqlDb := testSqlDB()
	v2db := v2.NewDB(sqlDb)
	v3db := NewDB(sqlDb)
	user := &model.Users{Name: "exql", Age: 6}
	err := resetTestDB(sqlDb)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("v2", func(b *testing.B) {
		for range b.N {
			_, err := v2db.Insert(user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("v3", func(b *testing.B) {
		for range b.N {
			_, err := v3db.Insert(user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkUpdate(b *testing.B) {
	sqlDb := testSqlDB()
	v2db := v2.NewDB(sqlDb)
	v3db := NewDB(sqlDb)
	user := &model.Users{Name: "exql", Age: 6}
	err := resetTestDB(sqlDb)
	if err != nil {
		b.Fatal(err)
	}
	_, err = v3db.Insert(user)
	if err != nil {
		b.Fatal(err)
	}
	userUpdate := &model.UpdateUsers{Name: Ptr("lqxe"), Age: Ptr(int64(8))}

	b.Run("v2", func(b *testing.B) {
		for range b.N {
			_, err := v2db.UpdateModel(userUpdate, v2.Where("id = ?", user.Id))
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("v3", func(b *testing.B) {
		for range b.N {
			_, err := v3db.UpdateModel(userUpdate, Where("id = ?", user.Id))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkMapFind(b *testing.B) {
	sqlDb := testSqlDB()
	v2db := v2.NewDB(sqlDb)
	v3db := NewDB(sqlDb)
	user := &model.Users{Name: "exql", Age: 6}
	err := resetTestDB(sqlDb)
	if err != nil {
		b.Fatal(err)
	}
	_, err = v3db.Insert(user)
	if err != nil {
		b.Fatal(err)
	}
	q := query.Q("SELECT * FROM `users` WHERE id = ?", user.Id)

	b.Run("v2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var user v2model.Users
			err := v2db.Find(q, &user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("v3", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var user model.Users
			err := v3db.Find(q, &user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkFindMany(b *testing.B) {
	sqlDb := testSqlDB()
	err := resetTestDB(sqlDb)
	if err != nil {
		b.Fatal(err)
	}
	var users []*model.Users
	for range 100 {
		users = append(users, &model.Users{Name: "exql", Age: 6})
	}

	v2db := v2.NewDB(sqlDb)
	v3db := NewDB(sqlDb)

	insertQuery, _ := v2.QueryForBulkInsert(users...)
	if _, err := v2db.Exec(insertQuery); err != nil {
		b.Fatal(err)
	}

	q := query.Q("SELECT * FROM `users` LIMIT ?", len(users))
	b.Run("v2", func(b *testing.B) {
		for range b.N {
			var user []*v2model.Users
			err := v2db.FindMany(q, &user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("v3", func(b *testing.B) {
		for range b.N {
			var user []*model.Users
			err := v3db.FindMany(q, &user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
