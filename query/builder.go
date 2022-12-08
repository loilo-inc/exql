package query

import (
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/xerrors"
)

type Query interface {
	Query() (string, []any, error)
}

type Insert struct {
	Into   string
	Values map[string]any
}

func backQuoteAndJoin(str ...string) string {
	var result []string
	for _, v := range str {
		result = append(result, fmt.Sprintf("`%s`", v))
	}
	return strings.Join(result, ",")
}

func (i *Insert) Query() (string, []any, error) {
	it := NewKeyIterator(i.Values)
	columns := backQuoteAndJoin(it.Keys()...)
	return fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES (%s)",
		i.Into, columns, SqlPlaceHolders(it.Size()),
	), it.Values(), nil
}

type InsertMany struct {
	Into    string
	Columns []string
	Values  [][]any
}

func (i *InsertMany) Query() (string, []any, error) {
	var values []string
	var args []any
	for _, v := range i.Values {
		if len(v) != len(i.Columns) {
			return "", nil, xerrors.Errorf("column value mismatch")
		}
		values = append(values, fmt.Sprintf("(%s)", SqlPlaceHolders(len(v))))
		args = append(args, v...)
	}
	return fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s",
		i.Into,
		backQuoteAndJoin(i.Columns...),
		strings.Join(values, ","),
	), args, nil
}

type Select struct {
	Columns []string
	From    string
	Where   Stmt
	Offset  int
	Limit   int
	OrderBy string
}

func (s *Select) Query() (string, []any, error) {
	if s.Where == nil {
		return "", nil, xerrors.Errorf("missing where statement")
	}
	where, err := s.Where.Query()
	if err != nil {
		return "", nil, err
	}
	var colmuns string
	if len(s.Columns) == 0 {
		colmuns = "*"
	} else {
		colmuns = backQuoteAndJoin(s.Columns...)
	}
	base := []string{fmt.Sprintf(
		"SELECT %s FROM `%s` WHERE %s",
		colmuns, s.From, where,
	)}
	var args []any
	args = append(args, s.Where.Args()...)
	appendOrderByLimitOffest(s, &base, &args)
	return strings.Join(base, " "), args, nil
}

type Update struct {
	Table   string
	Set     map[string]any
	Where   Stmt
	Limit   int
	Offset  int
	OrderBy string
}

func (q *Update) Query() (string, []any, error) {
	if q.Table == "" {
		return "", nil, xerrors.Errorf("empty table name")
	}
	valueQuery := QueryEx(q.Set)
	valueStmt, err := valueQuery.Query()
	if err != nil {
		return "", nil, err
	}
	whereStmt, err := q.Where.Query()
	if err != nil {
		return "", nil, err
	}
	var args []any
	args = append(args, valueQuery.Args()...)
	args = append(args, q.Where.Args()...)
	base := []string{fmt.Sprintf(
		"UPDATE `%s` SET %s WHERE %s",
		q.Table, valueStmt, whereStmt,
	)}
	appendOrderByLimitOffest(q, &base, &args)
	return strings.Join(base, " "), args, nil
}

type Delete struct {
	From    string
	Where   Stmt
	Limit   int
	Offset  int
	OrderBy string
}

func (d *Delete) Query() (string, []any, error) {
	where, err := d.Where.Query()
	if err != nil {
		return "", nil, err
	}
	base := []string{fmt.Sprintf("DELETE FROM `%s` WHERE %s", d.From, where)}
	var args []any
	args = append(args, d.Where.Args()...)
	appendOrderByLimitOffest(d, &base, &args)
	return strings.Join(base, " "), args, nil
}

func appendOrderByLimitOffest(p any, dest *[]string, args *[]any) {
	t := reflect.TypeOf(p).Elem()
	v := reflect.ValueOf(p).Elem()
	for i := 0; i < t.Elem().NumField(); i++ {
		f := t.Elem().Field(i)
		if f.Name == "OrderBy" {
			orderBy := v.Field(i).String()
			if orderBy != "" {
				*dest = append(*dest, fmt.Sprintf("ORDER BY %s", orderBy))
			}
		}
		if f.Name == "Limit" {
			limit := v.Field(i).Int()
			if limit > 0 {
				*dest = append(*dest, "LIMIT ?")
				*args = append(*args, limit)
			}
		}
		if f.Name == "Offset" {
			offset := v.Field(i).Int()
			if offset > 0 {
				*dest = append(*dest, "OFFSET ?")
				*args = append(*args, offset)
			}
		}
	}
}
