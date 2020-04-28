package exql

import (
	"database/sql"
	"fmt"
	"reflect"
)

type Mapper interface {
	// rowsから一行読んで構造体にマップする
	// pointerOfStructは必ず構造体へのポインタである必要がある
	// var user User
	// m.Map(rows, &user)
	Map(rows *sql.Rows, pointerOfStruct interface{}) error
	// rowsからすべての行を読んで構造体の配列にマップする
	// pointerOfSliceOfStructは必ず構造体のポインタのスライスへのポインタである必要がある
	// var users []*Users
	// m.MapMany(rows, &users)
	MapMany(rows *sql.Rows, pointerOfSliceOfStruct interface{}) error
}

type mapper struct {
}

func NewMapper() Mapper {
	return &mapper{}
}

type ColumnSplitter func(i int) string

type SerialMapper interface {
	// 結合されたrowから一行読んで構造体に順番にマップする
	// pointerOfStructは必ず構造体へのポインタの列挙である必要がある
	// var user User
	// var favorite UserFavorite
	// m.Map(rows, &user, &favorite)
	Map(rows *sql.Rows, pointersOfStruct ...interface{}) error
}

type serialMapper struct {
	splitter ColumnSplitter
}

func NewSerialMapper(s ColumnSplitter) SerialMapper {
	return &serialMapper{splitter: s}
}

func mapDestinationError() error {
	return fmt.Errorf("destination must be pointer of struct")
}
func (m *mapper) Map(row *sql.Rows, pointerOfStruct interface{}) error {
	if pointerOfStruct == nil {
		return mapDestinationError()
	}
	destValue := reflect.ValueOf(pointerOfStruct)
	destType := destValue.Type()
	if destType.Kind() != reflect.Ptr {
		return mapDestinationError()
	}
	destValue = destValue.Elem()
	if destValue.Kind() != reflect.Struct {
		return mapDestinationError()
	}
	if row.Next() {
		return mapRow(row, &destValue)
	} else {
		return fmt.Errorf("rows is empty")
	}
}

func mapManyDestinationError() error {
	return fmt.Errorf("destination must be pointer of slice of struct")

}
func (m *mapper) MapMany(rows *sql.Rows, structPtrOrSlicePtr interface{}) error {
	if structPtrOrSlicePtr == nil {
		return mapManyDestinationError()
	}
	destValue := reflect.ValueOf(structPtrOrSlicePtr)
	destType := destValue.Type()
	if destType.Kind() != reflect.Ptr {
		return mapManyDestinationError()
	}
	destType = destType.Elem()
	if destType.Kind() != reflect.Slice {
		return mapManyDestinationError()
	}
	// []*Model -> *Model
	sliceType := destType.Elem()
	if sliceType.Kind() != reflect.Ptr {
		return mapManyDestinationError()
	}
	// *Model -> Model
	sliceType = sliceType.Elem()
	for rows.Next() {
		// modelValue := SliceType{}
		modelValue := reflect.New(sliceType).Elem()
		if err := mapRow(rows, &modelValue); err != nil {
			return err
		}
		// *dest = append(*dest, i)
		destValue.Elem().Set(reflect.Append(destValue.Elem(), modelValue.Addr()))
	}
	return nil
}
func mapRow(
	row *sql.Rows,
	dest *reflect.Value,
) error {
	fields, err := aggregateFields(dest)
	if err != nil {
		return err
	}
	cols, err := row.ColumnTypes()
	if err != nil {
		return err
	}
	destVals := make([]interface{}, len(cols))
	for j, col := range cols {
		if fIndex, ok := fields[col.Name()]; ok {
			f := dest.Field(fIndex)
			destVals[j] = f.Addr().Interface()
		} else {
			ns := &noopScanner{}
			destVals[j] = ns
		}
	}
	return row.Scan(destVals...)
}

func aggregateFields(dest *reflect.Value) (map[string]int, error) {
	// *Model
	destType := dest.Type()
	fields := make(map[string]int)
	for i := 0; i < destType.NumField(); i++ {
		f := destType.Field(i)
		tag := f.Tag.Get("exql")
		if tag != "" {
			if f.Type.Kind() == reflect.Ptr {
				return nil, fmt.Errorf("struct field must not be pointer: %s %s", f.Type.Name(), f.Type.Kind())
			}
			tags, err := ParseTags(tag)
			if err != nil {
				return nil, err
			}
			col := tags["column"]
			fields[col] = i
		}
	}
	return fields, nil
}

func (s *serialMapper) Map(rows *sql.Rows, dest ...interface{}) error {
	var values []*reflect.Value
	for _, model := range dest {
		v := reflect.ValueOf(model)
		if v.Kind() != reflect.Ptr {
			return mapDestinationError()
		}
		v = v.Elem()
		if v.Kind() != reflect.Struct {
			return mapDestinationError()
		}
		values = append(values, &v)
	}
	if err := mapRowSerial(rows, values, s.splitter); err != nil {
		return err
	}
	return nil
}

func mapRowSerial(
	row *sql.Rows,
	destList []*reflect.Value,
	headColProvider ColumnSplitter,
) error {
	// *Model
	var destFields []map[string]int
	for _, dest := range destList {
		fields, err := aggregateFields(dest)
		if err != nil {
			return err
		}
		destFields = append(destFields, fields)
	}
	if len(destFields) == 0 {
		return fmt.Errorf("empty dest list")
	}
	cols, err := row.ColumnTypes()
	if err != nil {
		return err
	}
	destVals := make([]interface{}, len(cols))
	colIndex := 0
	for destIndex, dest := range destList {
		fields := destFields[destIndex]
		headCol := cols[colIndex]
		expectedHeadCol := headColProvider(destIndex)
		if headCol.Name() != expectedHeadCol {
			return fmt.Errorf(
				"head col mismatch: expected=%s, actual=%s",
				expectedHeadCol, headCol.Name(),
			)
		}
		start := colIndex
		ns := &noopScanner{}
		for ; colIndex < len(cols); colIndex++ {
			col := cols[colIndex]
			if colIndex > start && destIndex < len(destList)-1 {
				// Reach next column's head
				if col.Name() == headColProvider(destIndex+1) {
					break
				}
			}
			if fIndex, ok := fields[col.Name()]; ok {
				f := dest.Field(fIndex)
				destVals[colIndex] = f.Addr().Interface()
			} else {
				destVals[colIndex] = ns
			}
		}
	}
	return row.Scan(destVals...)
}

type noopScanner struct {
}

func (n *noopScanner) Scan(_ interface{}) error {
	// noop
	return nil
}
