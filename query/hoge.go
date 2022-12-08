package query

import (
	"fmt"
	"strings"

	"golang.org/x/xerrors"
)

type insertQuery struct {
	table   string
	entries KeyIterator
}

func Insert(table string, values map[string]any) Stmt {
	return &insertQuery{}
}

func (i *insertQuery) Query() (string, error) {
	var columns []string
	for _, v := range i.entries.Keys() {
		columns = append(columns, fmt.Sprintf("`%s`", v))
	}
	return fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES (%s)",
		i.table,
		strings.Join(columns, ","),
		SqlPlaceHolders(len(columns)),
	), nil
}

func (i *insertQuery) Args() []interface{} {
	return i.entries.Values()
}

type selectQuery struct {
	table  string
	where  Stmt
	offset int
	limit  int
}
type updateQuery struct {
	table  string
	values Stmt
	where  Stmt
}

func Update(table string, values Stmt, where Stmt) Stmt {
	return &updateQuery{
		table: table, values: values, where: where,
	}
}

func (q *updateQuery) Query() (string, error) {
	if q.table == "" {
		return "", xerrors.Errorf("empty table name")
	}
	valueStmt, err := q.values.Query()
	if err != nil {
		return "", err
	}
	whereStmt, err := q.where.Query()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"UPDATE `%s` SET %s WHERE %s",
		q.table, valueStmt, whereStmt,
	), nil
}

func (q *updateQuery) Args() []any {
	var args []any
	args = append(args, q.values.Args()...)
	args = append(args, q.where.Args()...)
	return args
}

type deleteQuery struct {
	table string
	where Stmt
}

func Delete(table string, where Stmt) Stmt {
	return &deleteQuery{
		table: table,
		where: where,
	}
}

func (d *deleteQuery) Query() (string, error) {
	where, err := d.where.Query()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("DELETE FROM `%s` WHERE %s", d.table, where), nil
}

func (d *deleteQuery) Args() []any {
	return d.where.Args()
}
