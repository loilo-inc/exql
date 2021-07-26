package exql

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
)

type Mapper interface {
	// Read single row and map columns to destination.
	// pointerOfStruct MUST BE a pointer of struct.
	// It closes rows after mapping regardless error occurred.
	// example:
	// 		var user User
	// 		err := m.Map(rows, &user)
	Map(rows *sql.Rows, pointerOfStruct interface{}) error
	// Read all rows and map columns for each destination.
	// pointerOfSliceOfStruct MUST BE a pointer of slices of pointer of struct.
	// It closes rows after mapping regardless error occurred.
	// example:
	// 		var users []*Users
	// 		m.MapMany(rows, &users)
	MapMany(rows *sql.Rows, pointerOfSliceOfStruct interface{}) error
}

type mapper struct {
}

func NewMapper() Mapper {
	return &mapper{}
}

// Error returned when record not found
var ErrRecordNotFound = errors.New("record not found")

type ColumnSplitter func(i int) string

type SerialMapper interface {
	// Read joined rows and map columns for each destination serially.
	// pointerOfStruct MUST BE a pointer of struct
	// NOTE: It WON'T close rows automatically. Close rows manually.
	// example:
	// 		var user User
	// 		var favorite UserFavorite
	// 		err := m.Map(rows, &user, &favorite)
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
	defer func() {
		if row != nil {
			row.Close()
		}
	}()
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
	}
	err := row.Close()
	if err != nil {
		return err
	}
	return ErrRecordNotFound
}

func mapManyDestinationError() error {
	return fmt.Errorf("destination must be pointer of slice of struct")

}
func (m *mapper) MapMany(rows *sql.Rows, structPtrOrSlicePtr interface{}) error {
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
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
	cnt := 0
	for rows.Next() {
		// modelValue := SliceType{}
		modelValue := reflect.New(sliceType).Elem()
		if err := mapRow(rows, &modelValue); err != nil {
			return err
		}
		// *dest = append(*dest, i)
		destValue.Elem().Set(reflect.Append(destValue.Elem(), modelValue.Addr()))
		cnt++
	}
	err := rows.Close()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return ErrRecordNotFound
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
	// *Model || **Model
	destType := dest.Type()
	if dest.Kind() == reflect.Ptr {
		destType = destType.Elem()
	}
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

type destKind uint

const (
	destKindInvalid destKind = iota
	destKindPtrStruct
	destKindPtrPtrStruct
)

var errMapRowSerialDestination = fmt.Errorf("destination must be either *struct or **struct")

func (s *serialMapper) Map(rows *sql.Rows, dest ...interface{}) error {
	var values []*reflect.Value
	var kind destKind
	for modelIdx, model := range dest {
		if modelIdx == 0 {
			kind = determineDestKind(reflect.TypeOf(model))
			if kind == destKindInvalid {
				return errMapRowSerialDestination
			}
		} else if kind != determineDestKind(reflect.TypeOf(model)) {
			return errMapRowSerialDestination
		}
		v := reflect.ValueOf(model).Elem()
		values = append(values, &v)
	}
	if kind == destKindPtrPtrStruct {
		if err := mapOuterJoinRowSerial(rows, values, s.splitter); err != nil {
			return err
		}
	} else {
		if err := mapRowSerial(rows, values, s.splitter); err != nil {
			return err
		}
	}
	return nil
}

func determineDestKind(destType reflect.Type) destKind {
	if destType == nil {
		return destKindInvalid
	} else if destType.Kind() != reflect.Ptr {
		return destKindInvalid
	} else if destType := destType.Elem(); destType.Kind() == reflect.Struct {
		return destKindPtrStruct
	} else if destType.Kind() != reflect.Ptr {
		return destKindInvalid
	} else if destType := destType.Elem(); destType.Kind() == reflect.Struct {
		return destKindPtrPtrStruct
	} else {
		return destKindInvalid
	}
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

// mapOuterJoinRowSerial maps rows to **Model. NULL value in ColumnSplitter column sets *Model = nil.
// It can map inner join rows as well but using mapRowSerial is faster because less reflection logics.
func mapOuterJoinRowSerial(
	row *sql.Rows,
	destList []*reflect.Value,
	headColProvider ColumnSplitter,
) error {
	var destFields []map[string]int
	destTypes := map[int]reflect.Type{}
	for destIndex, dest := range destList {
		fields, err := aggregateFields(dest)
		if err != nil {
			return err
		}
		destFields = append(destFields, fields)
		destTypes[destIndex] = dest.Type().Elem() // Model
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
	for destIndex := range destList {
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
		model := reflect.New(destTypes[destIndex]).Elem() // Model
		for ; colIndex < len(cols); colIndex++ {
			col := cols[colIndex]
			if colIndex > start && destIndex < len(destList)-1 {
				// Reach next column's head
				if col.Name() == headColProvider(destIndex+1) {
					break
				}
			}
			if fIndex, ok := fields[col.Name()]; ok {
				f := model.Field(fIndex)
				value := reflect.New(f.Addr().Type()) // **(Model.Field)
				destVals[colIndex] = value.Interface()
			} else {
				destVals[colIndex] = ns
			}
		}
	}
	if err := row.Scan(destVals...); err != nil {
		return err
	}

	colIndex = 0
	for destIndex, dest := range destList {
		if reflect.ValueOf(destVals[colIndex]).Elem().IsNil() { // *(Model.Field) == nil
			colIndex += len(destList)
			dest.Set(reflect.Zero(dest.Type()))
			continue
		}
		model := reflect.New(destTypes[destIndex]) // *Model
		fields := destFields[destIndex]
		start := colIndex
		for ; colIndex < len(cols); colIndex++ {
			col := cols[colIndex]
			if colIndex > start && destIndex < len(destList)-1 {
				// Reach next column's head
				if col.Name() == headColProvider(destIndex+1) {
					break
				}
			}
			if fIndex, ok := fields[col.Name()]; ok {
				f := model.Elem().Field(fIndex)
				f.Set(reflect.ValueOf(destVals[colIndex]).Elem().Elem())
			}
		}
		dest.Set(model)
	}

	return nil
}

type noopScanner struct {
}

func (n *noopScanner) Scan(_ interface{}) error {
	// noop
	return nil
}
