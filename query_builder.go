package exql

import (
	"fmt"
	"reflect"
	"strings"
)

type SaveQuery struct {
	Query           string
	Fields          []string
	Values          []interface{}
	PrimaryKeyField *reflect.Value
}

type QueryBuilder interface {
	Insert(structPtr interface{}) (*SaveQuery, error)
	Update(table string, set map[string]interface{}, where WhereQuery) (*SaveQuery, error)
}
type queryBuilder struct {
}

func (q *queryBuilder) Insert(modelPtr interface{}) (*SaveQuery, error) {
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
	hasPrimary := false
	var primaryKeyField reflect.Value
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
			if _, primary := tags["primary"]; primary {
				hasPrimary = true
				primaryKeyField = objValue.Elem().Field(i)
				// 主キーはVALUESに含めない
				continue
			}
			columns = append(columns, fmt.Sprintf(`%s`, colName))
			placeholders = append(placeholders, "?")
			values = append(values, objValue.Elem().Field(i).Interface())
		}
	}
	if !hasPrimary {
		return nil, fmt.Errorf("table has no primary key")
	}
	colCnt := len(columns)
	valCnt := len(values)
	if colCnt == 0 || valCnt == 0 {
		return nil, fmt.Errorf("obj doesn't have exql tags in any fields")
	}
	getTableName := objValue.MethodByName("TableName")
	if getTableName.IsNil() {
		return nil, fmt.Errorf("obj doesn't implement TableName() method")
	}
	tableName := getTableName.Call(nil)[0].String()
	if tableName == "" {
		return nil, fmt.Errorf("wrong implementation of TableName()")
	}
	query := fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
	return &SaveQuery{
		Query:           query,
		Fields:          columns,
		Values:          values,
		PrimaryKeyField: &primaryKeyField,
	}, nil
}

func (q *queryBuilder) Update(table string, set map[string]interface{}, where WhereQuery) (*SaveQuery, error) {
	if table == "" {
		return nil, fmt.Errorf("empty table name")
	}
	if len(set) == 0 {
		return nil, fmt.Errorf("empty field set")
	}
	var fields []string
	var assignments []string
	var values []interface{}
	for k, v := range set {
		f := fmt.Sprintf("`%s` = ?", k)
		assignments = append(assignments, f)
		fields = append(fields, k)
		values = append(values, v)
	}
	whereQ, err := where.Query()
	if err != nil {
		return nil, err
	}
	values = append(values, where.Args()...)
	query := fmt.Sprintf(
		"UPDATE `%s` SET %s WHERE %s",
		table, strings.Join(assignments, ", "), whereQ,
	)
	return &SaveQuery{
		Query:  query,
		Fields: fields,
		Values: values,
	}, nil
}
