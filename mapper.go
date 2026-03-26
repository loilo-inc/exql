package exql

import (
	"fmt"
	"reflect"
)

// Error returned when record not found
type ErrRecordNotFound struct{}

func (e ErrRecordNotFound) Error() string {
	return "record not found"
}

// ColumnSplitter is a function type for providing head column name for each destination struct in SerialMapper.
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
	Map(rows SqlRows, pointersOfStruct ...any) error
}

type serialMapper struct {
	splitter ColumnSplitter
}

func NewSerialMapper(s ColumnSplitter) SerialMapper {
	return &serialMapper{splitter: s}
}

var errMapDestination = fmt.Errorf("destination must be a pointer of struct")

// MapRow reads data from single row and maps those columns into destination struct.
// pointerOfStruct MUST BE a pointer of struct.
// It closes rows after mapping regardless error occurred.
//
// Example:
//
//	var user User
//	err := exql.MapRow(rows, &user)
func MapRow(
	row SqlRows,
	pointerOfStruct any,
) error {
	defer row.Close()

	destValue, err := resolveDestination(pointerOfStruct)
	if err != nil {
		return err
	}
	scanned := false
	if row.Next() {
		cols, err := row.Columns()
		if err != nil {
			return err
		}
		schema, err := parseMapSchema(destValue.Type())
		if err != nil {
			return err
		}
		receivers := schema.createReceivers(cols, destValue)
		if err := row.Scan(receivers...); err != nil {
			return err
		}
		scanned = true
	}
	if err := row.Err(); err != nil {
		return err
	} else if !scanned {
		return ErrRecordNotFound{}
	}
	return nil
}

// MapRows reads all data from rows and maps those columns for each destination struct.
// pointerOfSliceOfStruct MUST BE a pointer of slice of pointer of struct.
// It closes rows after mapping regardless error occurred.
//
// Example:
//
//	var users []*Users
//	err := exql.MapRows(rows, &users)
func MapRows(
	rows SqlRows,
	ptrOfSliceOfModelPtr any,
) error {
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	sliceType, destValue, err := resolveDestinationMany(ptrOfSliceOfModelPtr)
	if err != nil {
		return err
	}
	schema, err := parseMapSchema(sliceType)
	if err != nil {
		return err
	}
	cnt := 0
	for rows.Next() {
		// modelValue := SliceType{}
		modelValue := reflect.New(sliceType).Elem()
		receivers := schema.createReceivers(cols, &modelValue)
		if err := rows.Scan(receivers...); err != nil {
			return err
		}
		// *dest = append(*dest, i)
		destValue.Elem().Set(reflect.Append(destValue.Elem(), modelValue.Addr()))
		cnt++
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if cnt == 0 {
		return ErrRecordNotFound{}
	}
	return nil
}

func (m *serialMapper) Map(
	rows SqlRows,
	dest ...any,
) error {
	var values []*nullableDest

	if len(dest) == 0 {
		return fmt.Errorf("empty dest list")
	}

	for _, model := range dest {
		destValue, err := resolveNullableDestination(model)
		if err != nil {
			return err
		}
		values = append(values, destValue)
	}
	return mapJoinedRows(rows, values, m.splitter)
}

func mapJoinedRows(
	row SqlRows,
	destList []*nullableDest,
	headColProvider ColumnSplitter,
) error {
	// *Model || **Model
	var destFields []map[string]int
	destTypes := map[int]reflect.Type{}
	for destIndex, dest := range destList {
		md, err := parseMapSchema(dest.elemType)
		if err != nil {
			return err
		}
		destFields = append(destFields, md.fields)
		destTypes[destIndex] = dest.value.Type() // Model || *Model
	}
	cols, err := row.Columns()
	if err != nil {
		return err
	}
	destVals := make([]any, len(cols))
	colIndex := 0
	columnCounts := map[int]int{}
	for destIndex, dest := range destList {
		fields := destFields[destIndex]
		headCol := cols[colIndex]
		expectedHeadCol := headColProvider(destIndex)
		if headCol != expectedHeadCol {
			return fmt.Errorf(
				"head col mismatch: expected=%s, actual=%s",
				expectedHeadCol, headCol,
			)
		}
		start := colIndex
		ns := &noopScanner{}
		model := dest.value
		if destTypes[destIndex].Kind() == reflect.Pointer {
			m := reflect.New(destTypes[destIndex].Elem()).Elem() // Model
			model = &m
		}
		for ; colIndex < len(cols); colIndex++ {
			col := cols[colIndex]
			if colIndex > start && destIndex < len(destList)-1 {
				// Reach next column's head
				if col == headColProvider(destIndex+1) {
					columnCounts[destIndex] = colIndex - start
					break
				}
			} else if destIndex == len(destList)-1 {
				columnCounts[destIndex]++
			}
			if fIndex, ok := fields[col]; ok {
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
			if fIndex, ok := fields[col]; ok {
				f := model.Elem().Field(fIndex)
				if t := reflect.ValueOf(destVals[colIndex]).Elem(); t.IsNil() {
					f.Set(reflect.Zero(t.Type().Elem())) // To set (*null.Type)(nil) as null.Type{}
				} else {
					f.Set(reflect.ValueOf(destVals[colIndex]).Elem().Elem())
				}
			}
		}
		dest.value.Set(model) // dest = *Model
	}

	return nil
}

type noopScanner struct {
}

func (n *noopScanner) Scan(_ any) error {
	// noop
	return nil
}
