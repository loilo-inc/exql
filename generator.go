package exql

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type GenerateOptions struct {
	OutDir  string
	Package string
}

type templateData struct {
	Imports              string
	Model                string
	ModelLower           string
	M                    string
	Package              string
	Fields               string
	ScannedFields        string
	TableName            string
	HasPrimaryKey        bool
	PrimaryKeyFieldIndex int
}

func (d *db) Generate(opts *GenerateOptions) error {
	rows, err := d.db.Query(`show tables`)
	if err != nil {
		return err
	}
	if _, err := os.Stat(opts.OutDir); os.IsNotExist(err) {
		err := os.Mkdir(opts.OutDir, 0777)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return err
		}
		tables = append(tables, table)
	}
	for _, table := range tables {
		if err := d.generateModelFile(table, opts); err != nil {
			return err
		}
	}
	return nil
}

func (d *db) generateModelFile(tableName string, opt *GenerateOptions) error {
	tmpl := template.Must(template.New("model").Parse(modelTemplate))
	p := NewParser()
	table, err := p.ParseTable(d.db, tableName)
	if err != nil {
		return err
	}
	var imports []string
	if table.HasNullField() {
		imports = append(imports, `import "github.com/volatiletech/null"`)
	}
	if table.HasTimeField() {
		imports = append(imports, `import "time"`)
	}
	fields := strings.Builder{}
	scannedFields := strings.Builder{}
	for i, col := range table.Columns {
		scannedFields.WriteString(fmt.Sprintf(
			"\t\t&%s.%s,", table.TableName[0:1], col.Field()),
		)
		fields.WriteString(fmt.Sprintf("\t\t%s", col.Field()))
		if i < len(table.Columns)-1 {
			scannedFields.WriteString("\n")
			fields.WriteString("\n")
		}
	}
	data := &templateData{
		Imports:              strings.Join(imports, "\n"),
		Model:                strcase.ToCamel(table.TableName),
		ModelLower:           strcase.ToLowerCamel(table.TableName),
		M:                    table.TableName[0:1],
		Package:              opt.Package,
		Fields:               fields.String(),
		TableName:            tableName,
		ScannedFields:        scannedFields.String(),
		HasPrimaryKey:        table.HasPrimaryKey(),
		PrimaryKeyFieldIndex: table.PrimaryKeyFieldIndex(),
	}
	fileName := fmt.Sprintf("%s.go", strcase.ToSnake(table.TableName))
	outFile, err := os.Create(filepath.Join(opt.OutDir, fmt.Sprintf(fileName)))
	if err != nil {
		return err
	}
	defer outFile.Close()
	if err := tmpl.Execute(outFile, data); err != nil {
		return err
	}
	return nil
}

const modelTemplate = `// This file is generated by exql. DO NOT edit.
package {{.Package}}

{{.Imports}}

type {{.Model}} struct {
{{.Fields}}
}

func ({{.M}} *{{.Model}}) TableName() string {
    return "{{.TableName}}"
}

type {{.ModelLower}}Table struct {
}

var {{.Model}}Table = &{{.ModelLower}}Table{}

func ({{.M}} *{{.ModelLower}}Table) Name() string {
	return "{{.TableName}}"
}
`
