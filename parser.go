package exql

import (
	"database/sql"
	"fmt"
	"github.com/iancoleman/strcase"
	"regexp"
	"strings"
	"text/template"
)

type parser struct {
	tmpl   *template.Template
	outDir string
}

type Parser interface {
	ParseTable(db *sql.DB, table string) (*Table, error)
}

func NewParser() Parser {
	return &parser{}
}

type modelData struct {
	Package       string
	Model         string
	Models        string
	M             string
	Fields        string
	TableName     string
	ScannedFields string
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

type Column struct {
	FieldName    string         `json:"field_name"`
	FieldType    string         `json:"field_type"`
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
	empty := regexp.MustCompile("^\\s*$")
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

// `exql:"column:field_name;primary;auto_increment;"`
func (c *Column) Field() string {
	var tag []string
	tag = append(tag, fmt.Sprintf("column:%s", c.FieldName))
	if c.IsPrimary() {
		tag = append(tag, "primary")
	}
	tag = append(tag, c.ParseExtra()...)
	return fmt.Sprintf("%s %s `exql:\"%s\" json:\"%s\"`",
		strcase.ToCamel(c.FieldName),
		c.GoFieldType,
		strings.Join(tag, ";"),
		strcase.ToSnake(c.FieldType),
	)
}

func (p *parser) ParseTable(db *sql.DB, table string) (*Table, error) {
	rows, err := db.Query(fmt.Sprintf("show columns from %s", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []*Column
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
		parsedType, err := p.ParseType(_type, _null.String == "YES")
		if err != nil {
			return nil, err
		}
		cols = append(cols, &Column{
			FieldName:    field,
			FieldType:    field,
			GoFieldType:  parsedType,
			Nullable:     _null.String == "YES",
			DefaultValue: _default,
			Key:          key,
			Extra:        extra,
		})
	}
	return &Table{
		TableName: table,
		Columns:   cols,
	}, nil
}

func (p *parser) ParseType(t string, nullable bool) (string, error) {
	intPat := regexp.MustCompile("^(tiny|small|medium|big)?int\\(\\d+?\\)( unsigned)?( zerofill)?$")
	floatPat := regexp.MustCompile("^float$")
	doublePat := regexp.MustCompile("^double$")
	charPat := regexp.MustCompile("^(var)?char\\(\\d+?\\)$")
	textPat := regexp.MustCompile("^(tiny|medium|long)?text$")
	blobPat := regexp.MustCompile("^(tiny|medium|long)?blob$")
	datePat := regexp.MustCompile("^(date|datetime|timestamp|time)$")
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
			if unsigned {
				if !is64 {
					return "null.Uint", nil
				} else {
					return "null.Uint64", nil
				}
			} else {
				if !is64 {
					return "null.Int", nil
				} else {
					return "null.Int64", nil
				}
			}
		} else {
			if unsigned {
				if !is64 {
					return "uint", nil
				} else {
					return "uint64", nil
				}
			} else {
				if !is64 {
					return "int", nil
				} else {
					return "int64", nil
				}
			}
		}
	} else if datePat.MatchString(t) {
		if nullable {
			return "null.Time", nil
		}
		return "time.Time", nil
	} else if textPat.MatchString(t) || charPat.MatchString(t) {
		if nullable {
			return "null.String", nil
		}
		return "string", nil
	} else if floatPat.MatchString(t) {
		if nullable {
			return "null.Float32", nil
		}
		return "float32", nil
	} else if doublePat.MatchString(t) {
		if nullable {
			return "null.Float64", nil
		}
		return "float64", nil
	} else if blobPat.MatchString(t) {
		if nullable {
			return "null.Bytes", nil
		}
		return "[]byte", nil
	}
	return "", fmt.Errorf("unknown type: %s", t)
}
