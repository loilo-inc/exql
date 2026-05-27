package exql

import (
	"database/sql"
	"go/format"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTable_GenerateModelFile(t *testing.T) {
	table := &Table{
		TableName: "audit_logs",
		Columns: []*Column{
			{
				FieldName:   "id",
				FieldType:   "int(11)",
				GoFieldType: "int64",
				Key:         sql.NullString{String: "PRI", Valid: true},
				Extra:       sql.NullString{String: "auto_increment", Valid: true},
			},
			{
				FieldName:   "payload",
				FieldType:   "json",
				GoFieldType: "json.RawMessage",
			},
			{
				FieldName:   "created_at",
				FieldType:   "datetime",
				GoFieldType: "time.Time",
			},
			{
				FieldName:   "deleted_at",
				FieldType:   "datetime",
				GoFieldType: "null.Time",
				Nullable:    true,
			},
		},
	}

	file, err := table.GenerateModelFile("dist")
	assert.NoError(t, err)
	assert.Equal(t, "audit_logs.go", file.Name)

	fmted, err := format.Source(file.Source)
	assert.NoError(t, err)
	source := string(fmted)
	assert.Contains(t, source, "package dist")
	assert.Contains(t, source, `import "encoding/json"`)
	assert.Contains(t, source, `import "time"`)
	assert.Contains(t, source, `import "github.com/loilo-inc/exql/v3/null"`)
	assert.Contains(t, source, "type AuditLogs struct")
	assert.Regexp(t, regexp.MustCompile(`Payload\s+json\.RawMessage`), source)
	assert.Regexp(t, regexp.MustCompile(`CreatedAt\s+time\.Time`), source)
	assert.Regexp(t, regexp.MustCompile(`DeletedAt\s+null\.Time`), source)
	assert.Contains(t, source, `const AuditLogsTableName = "audit_logs"`)
}

func TestTable_GenerateModelFile_EscapesTableNameGoLiteral(t *testing.T) {
	tableName := `x";func init(){panic(1)};var _="`
	table := &Table{
		TableName: tableName,
		Columns: []*Column{
			{
				FieldName:   "id",
				FieldType:   "int(11)",
				GoFieldType: "int64",
				Key:         sql.NullString{String: "PRI", Valid: true},
			},
		},
	}

	file, err := table.GenerateModelFile("dist")
	assert.NoError(t, err)

	_, err = format.Source(file.Source)
	assert.NoError(t, err)
	assert.Contains(t, string(file.Source), `const XfuncInitpanic1varTableName = "x\";func init(){panic(1)};var _=\""`)
	assert.NotContains(t, string(file.Source), "\nfunc init()")
}

func TestTable_GenerateModelFile_RequiresTableName(t *testing.T) {
	_, err := (&Table{}).GenerateModelFile("dist")
	assert.ErrorIs(t, err, errTableNameEmpty)
}
