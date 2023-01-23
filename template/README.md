exql
---
[![codecov](https://codecov.io/gh/loilo-inc/exql/branch/master/graph/badge.svg?token=aGixN2xIMP)](https://codecov.io/gh/loilo-inc/exql)

Safe, strict and clear ORM for Go

## Introduction

exql is a simple ORM library for MySQL, written in Go. It is designed to work at the minimum for real software development. It has a few, limited but enough convenient functionalities of SQL database.
We adopted the data mapper model, not the active record. Records in the database are mapped into structs simply. Each model has no state and also no methods to modify itself and sync database records. You need to write bare SQL code for every operation you need except for a few cases.

exql is designed by focusing on safety and clearness in SQL usage. In other words, we never generate any SQL statements that are potentially dangerous or have ambiguous side effects across tables and the database.

It does:

- make insert/update query from model structs.
- map rows returned from the database into structs.
- map joined table into one or more structs.
- provide a safe syntax for the transaction.
- provide a framework to build dynamic SQL statements safely.
- generate model codes automatically from the database.

It DOESN'T

- make delete/update statements across the table.
- make unexpectedly slow select queries that don't use correct indices.
- modify any database settings, schemas and indices.

## Table of contents

- [exql](#exql)
- [Introduction](#introduction)
- [Table of contents](#table-of-contents)
- [Usage](#usage)
  - [Open database connection](#open-database-connection)
  - [Code Generation](#code-generation)
    - [Generate models from the database](#generate-models-from-the-database)
    - [Auto-Generated code](#auto-generated-code)
  - [Execute queries](#execute-queries)
    - [Insert](#insert)
    - [Update](#update)
  - [Transaction](#transaction)
  - [Map rows into structs](#map-rows-into-structs)
    - [Map rows](#map-rows)
    - [For joined table](#for-joined-table)
    - [For outer-joined table](#for-outer-joined-table)
  - [Use query builder](#use-query-builder)
- [License](#license)

## Usage

### Open database connection

```go
{{.Open}}
```

### Code Generation

#### Generate models from the database

```go
{{.GenerateModels}}
```

#### Auto-Generated code

```go
{{.AutoGenerateCode}}
```

### Execute queries

#### Insert

```go
{{.Insert}}
```

#### Update

```go
{{.Update}}
```

### Transaction

```go
{{.Tx}}
```

### Map rows into structs

#### Map rows

```go
{{.MapRows}}
```

#### For joined table

```go
{{.MapJoinedRows}}
```

#### For outer-joined table

```go
{{.MapOuterJoinedRows}}
```

### Use query builder

```go
{{.QueryBuilder}}
```

## License

MIT License / Copyright (c) 2020-Present LoiLo inc.

