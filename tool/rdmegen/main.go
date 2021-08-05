package main

import (
	"os"
	"text/template"
)

func main() {
	t := template.Must(template.ParseFiles("template/README.md"))
	data := map[string]string{
		"Open":               catFile("open.go"),
		"GenerateModels":     catFile("generator.go"),
		"Insert":             catFile("insert.go"),
		"Update":             catFile("update.go"),
		"MapRows":            catFile("mapper.go"),
		"MapJoinedRows":      catFile("serial_mapper.go"),
		"MapOuterJoinedRows": catFile("outer_join.go"),
		"Tx":                 catFile("tx.go"),
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
	s, err := os.ReadFile("example/" + f)
	if err != nil {
		panic(err)
	}
	return string(s)
}
