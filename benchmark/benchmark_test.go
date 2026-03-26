package benchmark

import (
	"testing"

	v2 "github.com/loilo-inc/exql/v2"
	v2model "github.com/loilo-inc/exql/v2/model"
	v3 "github.com/loilo-inc/exql/v3"
	v3model "github.com/loilo-inc/exql/v3/model"
	query "github.com/loilo-inc/exql/v3/query"
)

func BenchmarkQueryForInsert(b *testing.B) {
	v2user := v2model.Users{}
	v3user := v3model.Users{}
	b.Run("v2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v2.QueryForInsert(&v2user)
		}
	})
	b.Run("v3", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v3.QueryForInsert(&v3user)
		}
	})
}

func BenchmarkQueryForBulkInsert(b *testing.B) {
	var v2users []*v2model.Users
	var v3users []*v3model.Users
	for range 100 {
		v2users = append(v2users, &v2model.Users{})
		v3users = append(v3users, &v3model.Users{})
	}
	b.Run("v2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v2.QueryForBulkInsert(v2users...)
		}
	})
	b.Run("v3", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v3.QueryForBulkInsert(v3users...)
		}
	})
}

func BenchmarkQueryForUpdate(b *testing.B) {
	v2user := v2model.UpdateUsers{}
	v3user := v3model.UpdateUsers{}
	v2where := v2.Where("id = ?", 1)
	v3where := v3.Where("id = ?", 1)
	b.Run("v2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v2.QueryForUpdateModel(&v2user, v2where)
		}
	})
	b.Run("v3", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v3.QueryForUpdateModel(&v3user, v3where)
		}
	})
}

func BenchmarkInsert(b *testing.B) {
	sqlDb := testSqlDB(b)
	v2db := v2.NewDB(sqlDb)
	v3db := v3.NewDB(sqlDb)
	user := &v3model.Users{Name: "exql", Age: 6}
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
	sqlDb := testSqlDB(b)
	v2db := v2.NewDB(sqlDb)
	v3db := v3.NewDB(sqlDb)
	user := &v3model.Users{Name: "exql", Age: 6}
	err := resetTestDB(sqlDb)
	if err != nil {
		b.Fatal(err)
	}
	_, err = v3db.Insert(user)
	if err != nil {
		b.Fatal(err)
	}
	userUpdate := &v3model.UpdateUsers{Name: v3.Ptr("lqxe"), Age: v3.Ptr(int64(8))}

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
			_, err := v3db.UpdateModel(userUpdate, v3.Where("id = ?", user.Id))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkMapFind(b *testing.B) {
	sqlDb := testSqlDB(b)
	v2db := v2.NewDB(sqlDb)
	v3db := v3.NewDB(sqlDb)
	user := &v3model.Users{Name: "exql", Age: 6}
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
			var user v3model.Users
			err := v3db.Find(q, &user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkFindMany(b *testing.B) {
	sqlDb := testSqlDB(b)
	err := resetTestDB(sqlDb)
	if err != nil {
		b.Fatal(err)
	}
	var users []*v3model.Users
	for range 100 {
		users = append(users, &v3model.Users{Name: "exql", Age: 6})
	}

	v2db := v2.NewDB(sqlDb)
	v3db := v3.NewDB(sqlDb)

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
			var user []*v3model.Users
			err := v3db.FindMany(q, &user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
