package main

import (
	"os"
	"text/template"
)

func main() {
	t := template.Must(template.ParseFiles("template/README.md"))
	data := map[string]string{
		"Open":               catFile("example/open.go"),
		"GenerateModels":     catFile("example/generator.go"),
		"Insert":             catFile("example/insert.go"),
		"Update":             catFile("example/update.go"),
		"Delete":             catFile("example/delete.go"),
		"Other":              catFile("example/other.go"),
		"MapRows":            catFile("example/mapper.go"),
		"MapJoinedRows":      catFile("example/serial_mapper.go"),
		"MapOuterJoinedRows": catFile("example/outer_join.go"),
		"Tx":                 catFile("example/tx.go"),
		"QueryBuilder":       catFile("example/query_builder.go"),
		"AutoGenerateCode":   catFile("model/users.go"),
	}
	o, err := os.Create("README.md")
	if err != nil {
		panic(err)
	}
	if err := t.Execute(o, data); err != nil {
		panic(err)
	}
}

func catFile(f string) string {
	s, err := os.ReadFile(f)
	if err != nil {
		panic(err)
	}
	return string(s)
}
