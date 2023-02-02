package exql_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/mocks/mock_exql"
	"github.com/loilo-inc/exql/v2/model"
	"github.com/loilo-inc/exql/v2/query"
	"github.com/stretchr/testify/assert"
)

func TestPreparedExecutor(t *testing.T) {
	setup := func(t *testing.T, db exql.Executor) exql.StmtExecutor {
		pex := exql.NewStmtExecutor(db)
		t.Cleanup(func() {
			assert.Nil(t, pex.Close())
		})
		return pex
	}
	t.Run("integration", func(t *testing.T) {
		db := testDb()
		user1 := model.Users{Name: "go"}
		user2 := model.Users{Name: "lang"}
		err := db.Transaction(func(tx exql.Tx) error {
			pex := setup(t, tx.Tx())
			saver := exql.NewSaver(pex)
			for _, user := range []*model.Users{&user1, &user2} {
				if _, err := saver.Insert(user); err != nil {
					return err
				}
				t.Cleanup(func() {
					db.Delete(model.UsersTableName, exql.Where("id = ?", user.Id))
				})
			}
			return nil
		})
		assert.Nil(t, err)
		var list []*model.Users
		err = db.FindMany(query.Q(
			`select * from users where id in (?,?)`, user1.Id, user2.Id),
			&list,
		)
		assert.Nil(t, err)
	})
	t.Run("mock", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		ex := setup(t, db)
		qm := regexp.QuoteMeta
		insertQ := "insert into `users` (`name`) values (?)"
		selectQ := "select * from `users` where `name` = ?"
		stmt1 := mock.ExpectPrepare(qm(insertQ)).WillBeClosed()
		stmt1.ExpectExec().WithArgs("go").WillReturnResult(sqlmock.NewResult(0, 0))
		stmt1.ExpectExec().WithArgs("og").WillReturnResult(sqlmock.NewResult(0, 0))
		stmt2 := mock.ExpectPrepare(qm(selectQ)).WillBeClosed()
		stmt2.ExpectQuery().WithArgs("go").WillReturnRows(sqlmock.NewRows([]string{}))
		_, err = ex.Exec(insertQ, "go")
		assert.NoError(t, err)
		_, err = ex.Exec(insertQ, "og")
		assert.NoError(t, err)
		_, err = ex.Query(selectQ, "go")
		assert.NoError(t, err)
		err = ex.Close()
		assert.NoError(t, err)
	})

	t.Run("preparation error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		stmt := "stmt"
		testFunc := func(t *testing.T, body func(ex exql.StmtExecutor) (err error)) {
			mock := mock_exql.NewMockExecutor(ctrl)
			mock.EXPECT().PrepareContext(gomock.Any(), stmt).Return(nil, fmt.Errorf("err"))
			ex := exql.NewStmtExecutor(mock)
			err := body(ex)
			assert.EqualError(t, err, "err")
		}
		t.Run("Exec", func(t *testing.T) {
			testFunc(t, func(ex exql.StmtExecutor) (err error) {
				_, err = ex.Exec(stmt)
				return
			})
		})
		t.Run("Query", func(t *testing.T) {
			testFunc(t, func(ex exql.StmtExecutor) (err error) {
				_, err = ex.Query(stmt)
				return
			})
		})
	})
	t.Run("Prepare bypass to the inner executor", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		mock.ExpectPrepare("stmt").WillBeClosed()
		ex := exql.NewStmtExecutor(db)
		stmt, err := ex.Prepare("stmt")
		stmt.Close()
		assert.Nil(t, err)
	})
	t.Run("QueryRow bypass to the inner executor", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mock := mock_exql.NewMockExecutor(ctrl)
		mock.EXPECT().QueryRowContext(gomock.Any(), "stmt").Return(nil)
		ex := exql.NewStmtExecutor(mock)
		row := ex.QueryRow("stmt")
		assert.Nil(t, row)
	})
}
