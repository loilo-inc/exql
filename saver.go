package exql

import (
	"database/sql"
	"fmt"
	"github.com/apex/log"
	"reflect"
	"strings"
)

type saveQuery struct {
	query           string
	fields          []string
	values          []interface{}
	primaryKeyField *reflect.Value
}

func BuildInsertQuery(obj interface{}) (*saveQuery, error) {
	objValue := reflect.ValueOf(obj)
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
			tags := ParseTags(t)
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
	return &saveQuery{
		query:           query,
		fields:          columns,
		values:          values,
		primaryKeyField: &primaryKeyField,
	}, nil
}

func (d *db) Insert(modelPtr interface{}) (sql.Result, error) {
	s, err := BuildInsertQuery(modelPtr)
	if err != nil {
		return nil, err
	}
	result, err := d.db.Exec(s.query, s.values...)
	if err != nil {
		return nil, err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	kind := s.primaryKeyField.Kind()
	if kind == reflect.Int64 {
		s.primaryKeyField.Set(reflect.ValueOf(lid))
	} else if kind == reflect.Uint64 {
		s.primaryKeyField.Set(reflect.ValueOf(uint64(lid)))
	} else {
		log.Warn("primary key is not int64/uint64. assigning lastInsertedId is skipped")
	}
	return result, nil
}

func BuildUpdateQuery(table string, set map[string]interface{}, where WhereQuery) (*saveQuery, error) {
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
	return &saveQuery{
		query:  query,
		fields: fields,
		values: values,
	}, nil
}

func (d *db) Update(table string, set map[string]interface{}, where WhereQuery) (sql.Result, error) {
	s, err := BuildUpdateQuery(table, set, where)
	if err != nil {
		return nil, err
	}
	return d.db.Exec(s.query, s.values...)
}
