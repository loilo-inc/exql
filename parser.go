package exql

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
	"golang.org/x/xerrors"
)

type parser struct{}

type Parser interface {
	ParseTable(db *sql.DB, table string) (*Table, error)
}

func NewParser() Parser {
	return &parser{}
}

type Table struct {
	TableName string    `json:"table_name"`
	Columns   []*Column `json:"columns"`
}

func (t *Table) Fields() []string {
	var ret []string
	for _, c := range t.Columns {
		ret = append(ret, c.Field())
	}
	return ret
}
func (t *Table) HasNullField() bool {
	for _, c := range t.Columns {
		if c.Nullable {
			return true
		}
	}
	return false
}

func (t *Table) HasTimeField() bool {
	for _, c := range t.Columns {
		if c.GoFieldType == "time.Time" {
			return true
		}
	}
	return false
}

func (t *Table) HasJsonField() bool {
	for _, c := range t.Columns {
		if c.GoFieldType == "json.RawMessage" {
			return true
		}
	}
	return false
}

type Column struct {
	FieldName    string         `json:"field_name"`
	FieldType    string         `json:"field_type"`
	FieldIndex   int            `json:"field_index"`
	GoFieldType  string         `json:"go_field_type"`
	Nullable     bool           `json:"nullable"`
	DefaultValue sql.NullString `json:"default_value"`
	Key          sql.NullString `json:"key"`
	Extra        sql.NullString `json:"extra"`
}

func (c *Column) IsPrimary() bool {
	return c.Key.String == "PRI"
}

func (c *Column) ParseExtra() []string {
	comps := strings.Split(c.Extra.String, " ")
	empty := regexp.MustCompile(`^\s*$`)
	var ret []string
	for i := 0; i < len(comps); i++ {
		v := strings.Trim(comps[i], " ")
		if empty.MatchString(v) {
			continue
		}
		ret = append(ret, v)
	}
	return ret
}

func (c *Column) Field() string {
	return c.field(c.GoFieldType)
}

func (c *Column) UpdateField() string {
	return c.field("*" + c.GoFieldType)
}

func (c *Column) field(goFiledType string) string {
	var tag []string
	tag = append(tag, fmt.Sprintf("column:%s", c.FieldName))
	tag = append(tag, fmt.Sprintf("type:%s", c.FieldType))
	if c.IsPrimary() {
		tag = append(tag, "primary")
	}
	if !c.Nullable {
		tag = append(tag, "not null")
	}
	tag = append(tag, c.ParseExtra()...)
	return fmt.Sprintf("%s %s `exql:\"%s\" json:\"%s\"`",
		strcase.ToCamel(c.FieldName),
		goFiledType,
		strings.Join(tag, ";"),
		strcase.ToSnake(c.FieldName),
	)
}

func (p *parser) ParseTable(db *sql.DB, table string) (*Table, error) {
	rows, err := db.Query(fmt.Sprintf("show columns from %s", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []*Column
	i := 0
	for rows.Next() {
		field := ""
		_type := ""
		_null := sql.NullString{}
		key := sql.NullString{}
		_default := sql.NullString{}
		extra := sql.NullString{}
		if err := rows.Scan(&field, &_type, &_null, &key, &_default, &extra); err != nil {
			return nil, err
		}
		parsedType, err := ParseType(_type, _null.String == "YES")
		if err != nil {
			return nil, err
		}
		cols = append(cols, &Column{
			FieldName:    field,
			FieldType:    _type,
			FieldIndex:   i,
			GoFieldType:  parsedType,
			Nullable:     _null.String == "YES",
			DefaultValue: _default,
			Key:          key,
			Extra:        extra,
		})
		i++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &Table{
		TableName: table,
		Columns:   cols,
	}, nil
}

var (
	intPat    = regexp.MustCompile(`^(tiny|small|medium|big)?int(\(\d+?\))?( unsigned)?( zerofill)?$`)
	floatPat  = regexp.MustCompile(`^float$`)
	doublePat = regexp.MustCompile(`^double$`)
	charPat   = regexp.MustCompile(`^(var)?char\(\d+?\)$`)
	textPat   = regexp.MustCompile(`^(tiny|medium|long)?text$`)
	blobPat   = regexp.MustCompile(`^(tiny|medium|long)?blob$`)
	datePat   = regexp.MustCompile(`^(date|datetime|datetime\(\d\)|timestamp|timestamp\(\d\))$`)
	timePat   = regexp.MustCompile(`^(time|time\(\d\))$`)
	jsonPat   = regexp.MustCompile(`^json$`)
)

const (
	nullUint64Type  = "null.Uint64"
	nullInt64Type   = "null.Int64"
	uint64Type      = "uint64"
	int64Type       = "int64"
	nullFloat64Type = "null.Float64"
	float64Type     = "float64"
	nullFloat32Type = "null.Float32"
	float32Type     = "float32"
	nullTimeType    = "null.Time"
	timeType        = "time.Time"
	nullStrType     = "null.String"
	strType         = "string"
	nullBytesType   = "null.Bytes"
	bytesType       = "[]byte"
	nullJsonType    = "null.JSON"
	jsonType        = "json.RawMessage"
)

func ParseType(t string, nullable bool) (string, error) {
	if intPat.MatchString(t) {
		m := intPat.FindStringSubmatch(t)
		unsigned := strings.Contains(t, "unsigned")
		is64 := false
		if len(m) > 2 {
			switch m[1] {
			case "big":
				is64 = true
			default:
			}
		}
		if nullable {
			if unsigned && is64 {
				return nullUint64Type, nil
			} else {
				return nullInt64Type, nil
			}
		} else {
			if unsigned && is64 {
				return uint64Type, nil
			} else {
				return int64Type, nil
			}
		}
	} else if datePat.MatchString(t) {
		if nullable {
			return nullTimeType, nil
		}
		return timeType, nil
	} else if timePat.MatchString(t) {
		if nullable {
			return nullStrType, nil
		}
		return strType, nil
	} else if textPat.MatchString(t) || charPat.MatchString(t) {
		if nullable {
			return nullStrType, nil
		}
		return strType, nil
	} else if floatPat.MatchString(t) {
		if nullable {
			return nullFloat32Type, nil
		}
		return float32Type, nil
	} else if doublePat.MatchString(t) {
		if nullable {
			return nullFloat64Type, nil
		}
		return float64Type, nil
	} else if blobPat.MatchString(t) {
		if nullable {
			return nullBytesType, nil
		}
		return bytesType, nil
	} else if jsonPat.MatchString(t) {
		if nullable {
			return nullJsonType, nil
		}
		return jsonType, nil
	}
	return "", xerrors.Errorf("unknown type: %s", t)
}
