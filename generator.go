package exql

import (
	"database/sql"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"regexp"
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

var safeModelFileNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+\.go$`)

type modelFileOutput struct {
	path   string
	source []byte
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
	var outputs []*modelFileOutput
	seenPaths := map[string]string{}
	for _, table := range tables {
		output, err := d.generateModelFile(table, opts)
		if err != nil {
			return err
		}
		if prevTable, ok := seenPaths[output.path]; ok {
			return fmt.Errorf("duplicate generated model file %q for tables %q and %q", output.path, prevTable, table)
		}
		seenPaths[output.path] = table
		outputs = append(outputs, output)
	}
	for _, output := range outputs {
		if err := writeModelFile(output); err != nil {
			return err
		}
	}
	return nil
}

func (d *generator) generateModelFile(tableName string, opt *GenerateOptions) (*modelFileOutput, error) {
	p := NewParser()
	table, err := p.ParseTable(d.db, tableName)
	if err != nil {
		return nil, err
	}
	modelFile, err := table.GenerateModelFile(opt.Package)
	if err != nil {
		return nil, err
	}
	outFileName := modelFile.Name
	if mappedFileName, ok := opt.FileNameMap[tableName]; ok {
		if err := validateMappedModelFileName(mappedFileName); err != nil {
			return nil, err
		}
		outFileName = mappedFileName
	}
	return &modelFileOutput{
		path:   filepath.Join(opt.OutDir, outFileName),
		source: modelFile.Source,
	}, nil
}

func writeModelFile(output *modelFileOutput) error {
	if fmted, err := format.Source(output.source); err != nil {
		return err
	} else if err := os.WriteFile(output.path, fmted, 0640); err != nil {
		return err
	}
	log.Printf("generated file: %s", output.path)
	return nil
}

func validateMappedModelFileName(name string) error {
	if !safeModelFileNamePattern.MatchString(name) {
		return fmt.Errorf("invalid model file name %q: must match [A-Za-z0-9_-]+.go", name)
	}
	return nil
}
