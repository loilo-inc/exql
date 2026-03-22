package exql

import (
	"fmt"
	"reflect"

	q "github.com/loilo-inc/exql/v3/query"
)

type modelSchema struct {
	autoIncrementField *int
	primaryKeyFields   []int
	updatableFields    []int
	fields             map[string]int
	columns            map[int]string
	forUpdate          bool
}

type modelValue struct {
	tableName          string
	autoIncrementField *reflect.Value
	values             q.KeyIterator[any]
}

func aggregateFields(t reflect.Type, forUpdate bool) (*modelSchema, error) {
	fields := map[string]int{}
	columns := map[int]string{}
	exqlTagCount := 0
	var updatableFields []int
	var primaryKeyFields []int
	var autoIncrementField *int
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag, ok := f.Tag.Lookup("exql")
		if !ok {
			continue
		}

		if !forUpdate && f.Type.Kind() == reflect.Pointer {
			return nil, fmt.Errorf("field must not be a pointer: %s %s", f.Type.Name(), f.Type.Kind())
		} else if forUpdate && f.Type.Kind() != reflect.Pointer {
			return nil, fmt.Errorf("field must be a pointer: %s %s", f.Type.Name(), f.Type.Kind())
		}

		tags, err := ParseTags(tag)
		if err != nil {
			return nil, err
		}
		colName, ok := tags["column"]
		if !ok || colName == "" {
			return nil, fmt.Errorf("column tag is not set")
		}
		fields[colName] = i
		columns[i] = colName
		exqlTagCount++
		_, primary := tags["primary"]
		if primary {
			primaryKeyFields = append(primaryKeyFields, i)
		}
		_, autoIncrement := tags["auto_increment"]
		if autoIncrement {
			autoIncrementField = &i
		}
		if !autoIncrement {
			updatableFields = append(updatableFields, i)
		}
	}

	if exqlTagCount == 0 {
		return nil, fmt.Errorf("obj doesn't have exql tags in any fields")
	}

	if len(primaryKeyFields) == 0 {
		return nil, fmt.Errorf("table has no primary key")
	}

	return &modelSchema{
		autoIncrementField: autoIncrementField,
		primaryKeyFields:   primaryKeyFields,
		updatableFields:    updatableFields,
		fields:             fields,
		columns:            columns,
		forUpdate:          forUpdate,
	}, nil
}

func (ms *modelSchema) aggregateModelValue(
	modelPtr Model,
) (*modelValue, error) {
	res, err := ms.aggregateValue(modelPtr)
	if err != nil {
		return nil, err
	}
	res.tableName = modelPtr.TableName()
	return res, nil
}

func (ms *modelSchema) aggregateModelUpdateValue(
	modelPtr ModelUpdate,
) (*modelValue, error) {
	res, err := ms.aggregateValue(modelPtr)
	if err != nil {
		return nil, err
	}
	res.tableName = modelPtr.UpdateTableName()
	return res, nil
}

func (ms *modelSchema) aggregateValue(
	modelPtr any,
) (*modelValue, error) {
	if modelPtr == nil {
		return nil, errModelNil
	}
	objValue := reflect.ValueOf(modelPtr)
	objType := objValue.Type()
	if objType.Kind() != reflect.Pointer || objType.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("object must be pointer of struct")
	}
	// *User -> User
	objType = objType.Elem()
	var autoIncrementField *reflect.Value
	if ms.autoIncrementField != nil {
		f := objValue.Elem().Field(*ms.autoIncrementField)
		autoIncrementField = &f
	}
	var data = map[string]any{}
	for _, idx := range ms.updatableFields {
		f := objValue.Elem().Field(idx)
		col := ms.columns[idx]
		if !ms.forUpdate {
			data[col] = f.Interface()
		} else if !f.IsNil() {
			col := ms.columns[idx]
			data[col] = f.Elem().Interface()
		}
	}
	return &modelValue{
		autoIncrementField: autoIncrementField,
		values:             q.NewKeyIterator(data),
	}, nil
}
