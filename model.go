package exql

import (
	"fmt"
	"reflect"
)

type upsertModelSchema struct {
	autoIncrementField *int
	columns            []column
	forUpdate          bool
	modelType          reflect.Type
}

type column struct {
	index int
	name  string
}

type mapModelSchema struct {
	fields map[string]int
}

type modelValue struct {
	autoIncrementField *reflect.Value
	values             map[string]any
}

func parseUpsertSchema(t reflect.Type, forUpdate bool) (*upsertModelSchema, error) {
	if t.Kind() != reflect.Struct {
		return nil, errTypeNotStruct
	}
	exqlTagCount := 0
	var columns []column
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
		colName := tags["column"]
		if colName == "" {
			return nil, fmt.Errorf("column tag is not set")
		}
		exqlTagCount++
		_, autoIncrement := tags["auto_increment"]
		if autoIncrement {
			fType := f.Type
			if forUpdate {
				fType = fType.Elem()
			}
			kind := fType.Kind()
			if kind != reflect.Int64 && kind != reflect.Uint64 {
				return nil, fmt.Errorf("auto_increment field must be int64 or uint64")
			}
			autoIncrementField = &i
		}
		if !autoIncrement {
			columns = append(columns, column{index: i, name: colName})
		}
	}

	if exqlTagCount == 0 {
		return nil, fmt.Errorf("no exql tags in any fields")
	}

	return &upsertModelSchema{
		autoIncrementField: autoIncrementField,
		columns:            columns,
		forUpdate:          forUpdate,
		modelType:          t,
	}, nil
}

func parseMapSchema(t reflect.Type) (*mapModelSchema, error) {
	if t.Kind() != reflect.Struct {
		return nil, errTypeNotStruct
	}
	fields := map[string]int{}
	exqlTagCount := 0
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("exql")
		if tag == "" {
			continue
		}
		tags, err := ParseTags(tag)
		if err != nil {
			return nil, err
		}
		colName := tags["column"]
		if colName == "" {
			return nil, fmt.Errorf("column tag is not set")
		}
		fields[colName] = i
		exqlTagCount++
	}

	if exqlTagCount == 0 {
		return nil, fmt.Errorf("no exql tags in any fields")
	}

	return &mapModelSchema{fields: fields}, nil
}

func (ms *upsertModelSchema) aggregateValue(
	modelPtr any,
) (*modelValue, error) {
	if modelPtr == nil {
		return nil, errModelNil
	}
	// Must be a pointer of struct. Ensured by aggregateUpsertSchema.
	objValue := reflect.ValueOf(modelPtr).Elem()
	var autoIncrementField *reflect.Value
	if ms.autoIncrementField != nil {
		f := objValue.Field(*ms.autoIncrementField)
		autoIncrementField = &f
	}
	var data = map[string]any{}
	for _, v := range ms.columns {
		f := objValue.Field(v.index)
		if !ms.forUpdate {
			data[v.name] = f.Interface()
		} else if !f.IsNil() {
			data[v.name] = f.Elem().Interface()
		}
	}
	return &modelValue{
		autoIncrementField: autoIncrementField,
		values:             data,
	}, nil
}

func (ms *mapModelSchema) createReceivers(
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

var errTypeNotStruct = fmt.Errorf("type must be struct")
var errTableNameEmpty = fmt.Errorf("empty table name")
