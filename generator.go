package exql

import (
	"database/sql"
	"fmt"
	"github.com/iancoleman/strcase"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Generator interface {
	Generate() error
}

type generator struct {
	input    *Options
	template *template.Template
}

type Options struct {
	OutDir  string
	Url     string
	Package string
}

type TemplateData struct {
	Imports       string
	Model         string
	ModelLower    string
	M             string
	Package       string
	Fields        string
	ScannedFields string
	TableName     string
}

func NewGenerator(input *Options) Generator {
	tmplPath := "./templates/model.go.tmpl"
	tmpl := template.Must(template.ParseFiles(tmplPath))
	if _, err := os.Stat(input.OutDir); os.IsNotExist(err) {
		err := os.Mkdir(input.OutDir, 0777)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}
	return &generator{input: input, template: tmpl}
}

func (g *generator) Generate() error {
	db, err := sql.Open("mysql", g.input.Url)
	if err != nil {
		return err
	}
	defer db.Close()
	rows, err := db.Query(`show tables`)
	if err != nil {
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
		if err := g.generateTable(db, table); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) generateTable(db *sql.DB, tableName string) error {
	p := NewParser()
	table, err := p.ParseTable(db, tableName)
	if err != nil {
		return err
	}
	var imports []string
	if table.HasNullField() {
		imports = append(imports,`import "github.com/volatiletech/null"`)
	}
	if table.HasTimeField() {
		imports = append(imports,`import "time"`)
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
	data := &TemplateData{
		Imports:       strings.Join(imports, "\n"),
		Model:         strcase.ToCamel(table.TableName),
		ModelLower:    strcase.ToLowerCamel(table.TableName),
		M:             table.TableName[0:1],
		Package:       g.input.Package,
		Fields:        fields.String(),
		TableName:     tableName,
		ScannedFields: scannedFields.String(),
	}
	fileName := fmt.Sprintf("%s.go", strcase.ToSnake(table.TableName))
	outFile, err := os.Create(filepath.Join(g.input.OutDir, fmt.Sprintf(fileName)))
	if err != nil {
		return err
	}
	defer outFile.Close()
	if err := g.template.Execute(outFile, data); err != nil {
		return err
	}
	return nil
}
