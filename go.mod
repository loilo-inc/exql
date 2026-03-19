module github.com/loilo-inc/exql/v3

go 1.25

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/go-sql-driver/mysql v1.9.3
	github.com/iancoleman/strcase v0.3.0
	github.com/stretchr/testify v1.11.1
)

require (
	go.uber.org/mock v0.6.0
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da
)

replace github.com/volatiletech/inflect => github.com/aarondl/inflect v0.0.2

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
