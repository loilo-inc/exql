//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE
package query

import (
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/xerrors"
)

var errEmptyTable = xerrors.Errorf("empty table")
var errEmyptValues = xerrors.Errorf("empty values")
var errEmptyWhereClause = xerrors.Errorf("empty where clause")

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

func (i Insert) Validate() error {
	if i.Into == "" {
		return errEmptyTable
	}
	if len(i.Values) == 0 {
		return errEmyptValues
	}
	return nil
}

func (i Insert) Query() (string, []any, error) {
	if err := i.Validate(); err != nil {
		return "", nil, err
	}
	it := NewKeyIterator(i.Values)
	columns := backQuoteAndJoin(it.Keys()...)
	return fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES (%s)",
		i.Into, columns, Placeholders(it.Size()),
	), it.Values(), nil
}

type InsertMany struct {
	Into    string
	Columns []string
	Values  [][]any
}

func (i InsertMany) Validate() error {
	if i.Into == "" {
		return errEmptyTable
	}
	if len(i.Columns) == 0 || len(i.Values) == 0 {
		return errEmyptValues
	}
	return nil
}

func (i InsertMany) Query() (string, []any, error) {
	if err := i.Validate(); err != nil {
		return "", nil, err
	}
	var values []string
	var args []any
	for _, v := range i.Values {
		if len(i.Columns) != len(v) {
			return "", nil, xerrors.Errorf("number of columns/values mismatch")
		}
		values = append(values, fmt.Sprintf("(%s)", Placeholders(len(v))))
		args = append(args, v...)
	}
	return fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s",
		i.Into,
		backQuoteAndJoin(i.Columns...),
		strings.Join(values, ","),
	), args, nil
}

type Select struct {
	Columns   []string
	From      string
	Where     Condition
	OrderBy   string
	Limit     int
	Offset    int
	ForUpdate bool
}

func (s Select) Validate() error {
	if s.From == "" {
		return errEmptyTable
	}
	if s.Where == nil {
		return errEmptyWhereClause
	}
	return nil
}

func (s Select) Query() (string, []any, error) {
	if err := s.Validate(); err != nil {
		return "", nil, err
	}
	stmt, args, err := s.Where.Condition()
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
		colmuns, s.From, stmt,
	)}
	appendOrderByLimitOffest(s, &base, &args)
	if s.ForUpdate {
		base = append(base, "FOR UPDATE")
	}
	return strings.Join(base, " "), args, nil
}

type Update struct {
	Table   string
	Set     map[string]any
	Where   Condition
	OrderBy string
	Limit   int
	Offset  int
}

func (q Update) Validate() error {
	if q.Table == "" {
		return errEmptyTable
	}
	if len(q.Set) == 0 {
		return errEmyptValues
	}
	if q.Where == nil {
		return errEmptyWhereClause
	}
	return nil
}

func (q Update) Query() (string, []any, error) {
	if err := q.Validate(); err != nil {
		return "", nil, err
	}
	it := NewKeyIterator(q.Set)
	setExprs := make([]string, it.Size())
	for i, v := range it.Keys() {
		setExprs[i] = fmt.Sprintf("`%s` = ?", v)
	}
	setStmt := strings.Join(setExprs, ",")
	whereStmt, whereArgs, err := q.Where.Condition()
	if err != nil {
		return "", nil, err
	}
	var args []any
	args = append(args, it.Values()...)
	args = append(args, whereArgs...)
	base := []string{fmt.Sprintf(
		"UPDATE `%s` SET %s WHERE %s",
		q.Table, setStmt, whereStmt,
	)}
	appendOrderByLimitOffest(q, &base, &args)
	return strings.Join(base, " "), args, nil
}

type Delete struct {
	From    string
	Where   Condition
	OrderBy string
	Limit   int
	Offset  int
}

func (d Delete) Validate() error {
	if d.From == "" {
		return errEmptyTable
	}
	if d.Where == nil {
		return errEmptyWhereClause
	}
	return nil
}

func (d Delete) Query() (string, []any, error) {
	if err := d.Validate(); err != nil {
		return "", nil, err
	}
	stmt, args, err := d.Where.Condition()
	if err != nil {
		return "", nil, err
	}
	base := []string{fmt.Sprintf("DELETE FROM `%s` WHERE %s", d.From, stmt)}
	appendOrderByLimitOffest(d, &base, &args)
	return strings.Join(base, " "), args, nil
}

func appendOrderByLimitOffest(p any, dest *[]string, args *[]any) {
	t := reflect.TypeOf(p)
	v := reflect.ValueOf(p)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Name == "OrderBy" {
			orderBy := v.Field(i).String()
			if orderBy != "" {
				*dest = append(*dest, fmt.Sprintf("ORDER BY %s", orderBy))
			}
		}
		if f.Name == "Limit" {
			limit := v.Field(i).Interface().(int)
			if limit > 0 {
				*dest = append(*dest, "LIMIT ?")
				*args = append(*args, limit)
			}
		}
		if f.Name == "Offset" {
			offset := v.Field(i).Interface().(int)
			if offset > 0 {
				*dest = append(*dest, "OFFSET ?")
				*args = append(*args, offset)
			}
		}
	}
}
