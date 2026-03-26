module github.com/loilo-inc/exql/v3/benchmark

go 1.25

require (
	github.com/go-sql-driver/mysql v1.9.3
	github.com/loilo-inc/exql/v2 v2.2.2
	github.com/loilo-inc/exql/v3 v3.0.0
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/friendsofgo/errors v0.9.2 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/volatiletech/inflect v0.0.1 // indirect
	github.com/volatiletech/null v8.0.0+incompatible // indirect
	github.com/volatiletech/sqlboiler v3.7.1+incompatible // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
)

replace github.com/loilo-inc/exql/v3 => ../
