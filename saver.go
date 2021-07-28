package exql

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type SaveQuery struct {
	Query              string
	Fields             []string
	Values             []interface{}
	AutoIncrementField *reflect.Value
}

type Saver interface {
	Insert(structPtr interface{}) (sql.Result, error)
	InsertContext(ctx context.Context, structPtr interface{}) (sql.Result, error)
	QueryForInsert(structPtr interface{}) (*SaveQuery, error)
	Update(table string, set map[string]interface{}, where Clause) (sql.Result, error)
	UpdateContext(ctx context.Context, table string, set map[string]interface{}, where Clause) (sql.Result, error)
	QueryForUpdate(table string, set map[string]interface{}, where Clause) (*SaveQuery, error)
}

type Executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type saver struct {
	ex Executor
}

type SET map[string]interface{}

func NewSaver(ex Executor) *saver {
	return &saver{ex: ex}
}

func (s *saver) Insert(modelPtr interface{}) (sql.Result, error) {
	return s.InsertContext(context.Background(), modelPtr)
}

func (s *saver) InsertContext(ctx context.Context, modelPtr interface{}) (sql.Result, error) {
	q, err := s.QueryForInsert(modelPtr)
	if err != nil {
		return nil, err
	}
	result, err := s.ex.ExecContext(ctx, q.Query, q.Values...)
	if err != nil {
		return nil, err
	}
	if q.AutoIncrementField != nil {
		lid, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		kind := q.AutoIncrementField.Kind()
		if kind == reflect.Int64 {
			q.AutoIncrementField.Set(reflect.ValueOf(lid))
		} else if kind == reflect.Uint64 {
			q.AutoIncrementField.Set(reflect.ValueOf(uint64(lid)))
		}
	}
	return result, nil
}

func (s *saver) Update(
	table string,
	set map[string]interface{},
	where Clause,
) (sql.Result, error) {
	return s.UpdateContext(context.Background(), table, set, where)
}

func (s *saver) UpdateContext(
	ctx context.Context,
	table string,
	set map[string]interface{},
	where Clause,
) (sql.Result, error) {
	q, err := s.QueryForUpdate(table, set, where)
	if err != nil {
		return nil, err
	}
	return s.ex.ExecContext(ctx, q.Query, q.Values...)
}

func (s *saver) QueryForInsert(modelPtr interface{}) (*SaveQuery, error) {
	if modelPtr == nil {
		return nil, fmt.Errorf("pointer is nil")
	}
	objValue := reflect.ValueOf(modelPtr)
	objType := objValue.Type()
	if objType.Kind() != reflect.Ptr || objType.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("object must be pointer of struct")
	}
	var columns []string
	var values []interface{}
	var placeholders []string
	// *User -> User
	objType = objType.Elem()
	exqlTagCount := 0
	var primaryKeyFields []*reflect.Value
	var autoIncrementField *reflect.Value
	for i := 0; i < objType.NumField(); i++ {
		f := objType.Field(i)
		if t, ok := f.Tag.Lookup("exql"); ok {
			tags, err := ParseTags(t)
			if err != nil {
				return nil, err
			}
			colName, ok := tags["column"]
			if !ok || colName == "" {
				return nil, fmt.Errorf("column tag is not set")
			}
			exqlTagCount++
			if _, primary := tags["primary"]; primary {
				primaryKeyField := objValue.Elem().Field(i)
				primaryKeyFields = append(primaryKeyFields, &primaryKeyField)
			}
			if _, autoIncrement := tags["auto_increment"]; autoIncrement {
				field := objValue.Elem().Field(i)
				autoIncrementField = &field
				// Not include auto_increment field in insert query
				continue
			}
			columns = append(columns, fmt.Sprintf("`%s`", colName))
			placeholders = append(placeholders, "?")
			values = append(values, objValue.Elem().Field(i).Interface())
		}
	}
	if exqlTagCount == 0 {
		return nil, fmt.Errorf("obj doesn't have exql tags in any fields")
	}

	if len(primaryKeyFields) == 0 {
		return nil, fmt.Errorf("table has no primary key")
	}

	getTableName := objValue.MethodByName("TableName")
	if !getTableName.IsValid() {
		return nil, fmt.Errorf("obj doesn't implement TableName() method")
	}
	tableName := getTableName.Call(nil)[0]
	if tableName.Type().Kind() != reflect.String {
		return nil, fmt.Errorf("wrong implementation of TableName()")
	}
	query := fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
	return &SaveQuery{
		Query:              query,
		Fields:             columns,
		Values:             values,
		AutoIncrementField: autoIncrementField,
	}, nil
}

type assignment struct {
	expression string
	field      string
	value      interface{}
}

func (s *saver) QueryForUpdate(table string, set map[string]interface{}, where Clause) (*SaveQuery, error) {
	if table == "" {
		return nil, fmt.Errorf("empty table name")
	}
	if len(set) == 0 {
		return nil, fmt.Errorf("empty field set")
	}
	if where.Type() != ClauseTypeWhere {
		return nil, fmt.Errorf("where is not build by Where()")
	}
	whereQ, err := where.Query()
	if err != nil {
		return nil, err
	}
	var assignments []*assignment
	for k, v := range set {
		f := fmt.Sprintf("`%s` = ?", k)
		assignments = append(assignments, &assignment{
			expression: f,
			field:      k,
			value:      v,
		})
	}
	sort.Slice(assignments, func(i, j int) bool {
		return strings.Compare(assignments[i].field, assignments[j].field) < 0
	})
	var fields []string
	var values []interface{}
	var expressions []string
	for _, v := range assignments {
		fields = append(fields, v.field)
		values = append(values, v.value)
		expressions = append(expressions, v.expression)
	}
	values = append(values, where.Args()...)
	query := fmt.Sprintf(
		"UPDATE `%s` SET %s WHERE %s",
		table, strings.Join(expressions, ", "), whereQ,
	)
	return &SaveQuery{
		Query:  query,
		Fields: fields,
		Values: values,
	}, nil
}
