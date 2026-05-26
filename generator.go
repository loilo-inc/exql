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
	// Values must be file names only: no absolute paths, path separators, "." or "..".
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
	if name == "" {
		return fmt.Errorf("invalid model file name %q: empty file name", name)
	}
	if filepath.IsAbs(name) {
		return fmt.Errorf("invalid model file name %q: absolute paths are not allowed", name)
	}
	if name == "." || name == ".." {
		return fmt.Errorf("invalid model file name %q: path traversal is not allowed", name)
	}
	if strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("invalid model file name %q: path separators are not allowed", name)
	}
	return nil
}
