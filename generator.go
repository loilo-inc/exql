package exql

import (
	"database/sql"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Generator interface {
	Generate(opts *GenerateOptions) error
}
type generator struct {
	db *sql.DB
}
type GenerateOptions struct {
	OutDir  string
	Package string
	Exclude []string
	// FileNameMap maps table names to output file names.
	// Values must match [A-Za-z0-9_-]+.go.
	FileNameMap map[string]string
}

func NewGenerator(db *sql.DB) Generator {
	return &generator{db: db}
}

func (d *generator) Generate(opts *GenerateOptions) error {
	rows, err := d.db.Query(`show tables`)
	if err != nil {
		return err
	}
	if opts.OutDir == "" {
		opts.OutDir = "model"
	}
	if opts.Package == "" {
		opts.Package = "model"
	}
	if _, err := os.Stat(opts.OutDir); os.IsNotExist(err) {
		err := os.Mkdir(opts.OutDir, 0750)
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
		for _, e := range opts.Exclude {
			if e == table {
				goto EOL
			}
		}
		tables = append(tables, table)
	EOL:
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, table := range tables {
		if err := d.generateModelFile(table, opts); err != nil {
			return err
		}
	}
	return nil
}

func (d *generator) generateModelFile(tableName string, opt *GenerateOptions) error {
	p := NewParser()
	table, err := p.ParseTable(d.db, tableName)
	if err != nil {
		return err
	}
	modelFile, err := table.GenerateModelFile(opt.Package)
	if err != nil {
		return err
	}
	outFileName := modelFile.Name
	if mappedFileName, ok := opt.FileNameMap[tableName]; ok {
		if err := validateMappedModelFileName(mappedFileName); err != nil {
			return err
		}
		outFileName = mappedFileName
	}
	outFile := filepath.Join(
		opt.OutDir,
		outFileName,
	)
	if fmted, err := format.Source(modelFile.Source); err != nil {
		return err
	} else if err := os.WriteFile(outFile, fmted, 0640); err != nil {
		return err
	}
	log.Printf("generated file: %s", outFile)
	return nil
}

func validateMappedModelFileName(name string) error {
	if !strings.HasSuffix(name, ".go") {
		return fmt.Errorf("invalid model file name %q: must match [A-Za-z0-9_-]+.go", name)
	}
	base := strings.TrimSuffix(name, ".go")
	if base == "" {
		return fmt.Errorf("invalid model file name %q: must match [A-Za-z0-9_-]+.go", name)
	}
	for _, r := range base {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_' ||
			r == '-' {
			continue
		}
		return fmt.Errorf("invalid model file name %q: must match [A-Za-z0-9_-]+.go", name)
	}
	return nil
}
