package exql

import (
	"fmt"
	"reflect"

	q "github.com/loilo-inc/exql/v3/query"
)

type modelSchema struct {
	autoIncrementField *int
	updatableFields    []int
	fields             map[string]int
	columns            map[int]string
	forUpdate          bool
	key                string
}

type modelValue struct {
	autoIncrementField *reflect.Value
	values             q.KeyIterator[any]
}

func aggregateFields(t reflect.Type, forUpdate bool) (*modelSchema, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type must be struct")
	}
	fields := map[string]int{}
	columns := map[int]string{}
	exqlTagCount := 0
	var updatableFields []int
	var autoIncrementField *int
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("exql")
		if tag == "" {
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
		_, autoIncrement := tags["auto_increment"]
		if autoIncrement {
			fType := f.Type
			if forUpdate {
				fType = fType.Elem()
			}
			if fType.Kind() != reflect.Int64 {
				return nil, fmt.Errorf("auto_increment field must be int64")
			}
			autoIncrementField = &i
		}
		if !autoIncrement {
			updatableFields = append(updatableFields, i)
		}
	}

	if exqlTagCount == 0 {
		return nil, fmt.Errorf("no exql tags in any fields")
	}

	return &modelSchema{
		autoIncrementField: autoIncrementField,
		updatableFields:    updatableFields,
		fields:             fields,
		columns:            columns,
		forUpdate:          forUpdate,
		key:                typeKey(t),
	}, nil
}

var errTableNameEmpty = fmt.Errorf("empty table name")

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
	if ms.key != typeKey(objType) {
		return nil, fmt.Errorf("model type mismatch: expected=%s, actual=%s", ms.key, typeKey(objType))
	}
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

func (ms *modelSchema) createReceivers(
	cols []string,
	dest *reflect.Value,
) []any {
	destVals := make([]any, len(cols))
	for j, col := range cols {
		if fIndex, ok := ms.fields[col]; ok {
			f := dest.Field(fIndex)
			destVals[j] = f.Addr().Interface()
		} else {
			ns := &noopScanner{}
			destVals[j] = ns
		}
	}
	return destVals
}
