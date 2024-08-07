package exql

import (
	"database/sql"
	"errors"
	"reflect"

	"golang.org/x/xerrors"
)

// Error returned when record not found
var ErrRecordNotFound = errors.New("record not found")

// Deprecated: Use Finder It will be removed in next version.
type Mapper interface {
	// Deprecated: Use Find or MapRow. It will be removed in next version.
	Map(rows *sql.Rows, destPtr any) error
	// Deprecated: Use FindContext or MapRows. It will be removed in next version.
	MapMany(rows *sql.Rows, destSlicePtr any) error
}

type mapper struct{}

// Map reads data from single row and maps those columns into destination struct.
func (m *mapper) Map(rows *sql.Rows, destPtr any) error {
	return MapRow(rows, destPtr)
}

// MapMany reads all data from rows and maps those columns for each destination struct.
func (m *mapper) MapMany(rows *sql.Rows, destSlicePtr any) error {
	return MapRows(rows, destSlicePtr)
}

type ColumnSplitter func(i int) string

// SerialMapper is an interface for mapping a joined row into one or more destinations serially.
type SerialMapper interface {
	// Map reads joined rows and maps columns for each destination serially.
	// The second argument, pointerOfStruct, MUST BE a pointer of the struct.
	//
	// NOTE: DO NOT FORGET to close rows manually, as it WON'T do it automatically.
	//
	// Example:
	//
	//	var user User
	//	var favorite UserFavorite
	//	defer rows.Close()
	//	err := m.Map(rows, &user, &favorite)
	Map(rows *sql.Rows, pointersOfStruct ...any) error
}

type serialMapper struct {
	splitter ColumnSplitter
}

func NewSerialMapper(s ColumnSplitter) SerialMapper {
	return &serialMapper{splitter: s}
}

var errMapDestination = xerrors.Errorf("destination must be a pointer of struct")

// MapRow reads data from single row and maps those columns into destination struct.
// pointerOfStruct MUST BE a pointer of struct.
// It closes rows after mapping regardless error occurred.
//
// Example:
//
//	var user User
//	err := exql.MapRow(rows, &user)
func MapRow(row *sql.Rows, pointerOfStruct interface{}) error {
	defer func() {
		if row != nil {
			row.Close()
		}
	}()
	if pointerOfStruct == nil {
		return errMapDestination
	}
	destValue := reflect.ValueOf(pointerOfStruct)
	destType := destValue.Type()
	if destType.Kind() != reflect.Ptr {
		return errMapDestination
	}
	destValue = destValue.Elem()
	if destValue.Kind() != reflect.Struct {
		return errMapDestination
	}
	if row.Next() {
		return mapRow(row, &destValue)
	}
	if err := row.Err(); err != nil {
		return err
	}
	err := row.Close()
	if err != nil {
		return err
	}
	return ErrRecordNotFound
}

var errMapManyDestination = xerrors.Errorf("destination must be a pointer of slice of struct")

// MapRows reads all data from rows and maps those columns for each destination struct.
// pointerOfSliceOfStruct MUST BE a pointer of slice of pointer of struct.
// It closes rows after mapping regardless error occurred.
//
// Example:
//
//	var users []*Users
//	err := exql.MapRows(rows, &users)
func MapRows(rows *sql.Rows, structPtrOrSlicePtr interface{}) error {
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	if structPtrOrSlicePtr == nil {
		return errMapManyDestination
	}
	destValue := reflect.ValueOf(structPtrOrSlicePtr)
	destType := destValue.Type()
	if destType.Kind() != reflect.Ptr {
		return errMapManyDestination
	}
	destType = destType.Elem()
	if destType.Kind() != reflect.Slice {
		return errMapManyDestination
	}
	// []*Model -> *Model
	sliceType := destType.Elem()
	if sliceType.Kind() != reflect.Ptr {
		return errMapManyDestination
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
	if err := rows.Err(); err != nil {
		return err
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
				return nil, xerrors.Errorf("struct field must not be a pointer: %s %s", f.Type.Name(), f.Type.Kind())
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

var errMapRowSerialDestination = xerrors.Errorf("destination must be either *(struct) or *((*struct)(nil))")

func (s *serialMapper) Map(rows *sql.Rows, dest ...interface{}) error {
	var values []*reflect.Value

	if len(dest) == 0 {
		return xerrors.Errorf("empty dest list")
	}

	for _, model := range dest {
		v := reflect.ValueOf(model)
		if v.Kind() != reflect.Ptr {
			return errMapRowSerialDestination
		}
		v = v.Elem()
		if v.Kind() == reflect.Struct {
			values = append(values, &v)
		} else if v.Kind() != reflect.Ptr {
			return errMapRowSerialDestination
		} else if !v.IsNil() || v.Type().Elem().Kind() != reflect.Struct {
			return errMapRowSerialDestination
		} else {
			values = append(values, &v)

		}
	}
	return mapRowSerial(rows, values, s.splitter)
}

func mapRowSerial(
	row *sql.Rows,
	destList []*reflect.Value,
	headColProvider ColumnSplitter,
) error {
	// *Model || **Model
	var destFields []map[string]int
	destTypes := map[int]reflect.Type{}
	for destIndex, dest := range destList {
		fields, err := aggregateFields(dest)
		if err != nil {
			return err
		}
		destFields = append(destFields, fields)
		destTypes[destIndex] = dest.Type() // Model || *Model
	}
	cols, err := row.ColumnTypes()
	if err != nil {
		return err
	}
	destVals := make([]interface{}, len(cols))
	colIndex := 0
	columnCounts := map[int]int{}
	for destIndex, dest := range destList {
		fields := destFields[destIndex]
		headCol := cols[colIndex]
		expectedHeadCol := headColProvider(destIndex)
		if headCol.Name() != expectedHeadCol {
			return xerrors.Errorf(
				"head col mismatch: expected=%s, actual=%s",
				expectedHeadCol, headCol.Name(),
			)
		}
		start := colIndex
		ns := &noopScanner{}
		model := dest
		if destTypes[destIndex].Kind() == reflect.Ptr {
			m := reflect.New(destTypes[destIndex].Elem()).Elem() // Model
			model = &m
		}
		for ; colIndex < len(cols); colIndex++ {
			col := cols[colIndex]
			if colIndex > start && destIndex < len(destList)-1 {
				// Reach next column's head
				if col.Name() == headColProvider(destIndex+1) {
					columnCounts[destIndex] = colIndex - start
					break
				}
			} else if destIndex == len(destList)-1 {
				columnCounts[destIndex]++
			}
			if fIndex, ok := fields[col.Name()]; ok {
				f := model.Field(fIndex)
				if destTypes[destIndex].Kind() == reflect.Struct {
					destVals[colIndex] = f.Addr().Interface() // *(Model.Field)
				} else {
					destVals[colIndex] = reflect.New(f.Addr().Type()).Interface() // **(Model.Field)
				}
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
		fields := destFields[destIndex]
		if destTypes[destIndex].Kind() == reflect.Struct || reflect.ValueOf(destVals[colIndex]).Elem().IsNil() {
			if destIndex < len(destList)-1 {
				colIndex += columnCounts[destIndex]
			}
			continue
		}

		model := reflect.New(destTypes[destIndex].Elem()) // *Model
		start := colIndex
		for ; colIndex < start+columnCounts[destIndex]; colIndex++ {
			col := cols[colIndex]
			if fIndex, ok := fields[col.Name()]; ok {
				f := model.Elem().Field(fIndex)
				if t := reflect.ValueOf(destVals[colIndex]).Elem(); t.IsNil() {
					f.Set(reflect.Zero(t.Type().Elem())) // To set (*null.Type)(nil) as null.Type{}
				} else {
					f.Set(reflect.ValueOf(destVals[colIndex]).Elem().Elem())
				}
			}
		}
		dest.Set(model) // dest = *Model
	}

	return nil
}

type noopScanner struct {
}

func (n *noopScanner) Scan(_ interface{}) error {
	// noop
	return nil
}
