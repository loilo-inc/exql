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
  - [Execute queries](#execute-queries)
    - [Insert](#insert)
    - [Update](#update)
    - [Delete](#delete)
    - [Other](#other)
  - [Transaction](#transaction)
  - [Find records](#find-records)
    - [For simple query](#for-simple-query)
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
exql provides an automated code generator of models based on the database schema. This is a typical table schema of MySQL database.

```
mysql> show columns from users;
+-------+--------------+------+-----+---------+----------------+
| Field | Type         | Null | Key | Default | Extra          |
+-------+--------------+------+-----+---------+----------------+
| id    | int(11)      | NO   | PRI | NULL    | auto_increment |
| name  | varchar(255) | NO   |     | NULL    |                |
| age   | int(11)      | NO   |     | NULL    |                |
+-------+--------------+------+-----+---------+----------------+
```

To generate model codes, based on the schema, you need to write the code like this:

```go
{{.GenerateModels}}
```

And results are mostly like this:

```go
{{.AutoGenerateCode}}
```

`Users` is the destination of the data mapper. It only has value fields and one method, `TableName()`. This is the implementation of `exql.Model` that can be passed into data saver. All structs, methods and field tags must be preserved as it is, for internal use. If you want to modify the results, you must run the generator again.

`UpdateUsers` is a partial structure for the data model. It has identical name fields to `Users`, but all types are represented as a pointer. It is used to update table columns partially. In other words, it is a designated, typesafe map for the model.

### Execute queries

There are several ways to publish SQL statements with exql.

#### Insert

INSERT query is constructed automatically based on model data and executed without writing the statement. To insert new records into the database, set values to the model and pass it to `exql.DB#Insert` method.

```go
{{.Insert}}
```

#### Update

UPDATE query is constructed automatically based on the model update struct. To avoid unexpected updates to the table, all values are represented by a pointer of data type.

```go
{{.Update}}
```

#### Delete

DELETE query is published to the table with given conditions. There's no way to construct DELETE query from the model as a security reason.

```go
{{.Delete}}
```

#### Other

Other queries should be executed by `sql.DB` that got from `DB`.

```go
{{.Other}}
```

### Transaction

Transaction with `BEGIN`~`COMMIT`/`ROLLBACK` is done by `TransactionWithContext`. You don't need to call `BeginTx` and `Commit`/`Rollback` manually and all atomic operations are done within a callback.

```go
{{.Tx}}
```

### Find records

To find records from the database, use `Find`/`FindMany` method. It executes the query and maps results into structs correctly.

#### For simple query

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

`exql/query` package is a low-level API for building complicated SQL statements. See [V2 Release Notes](https://github.com/loilo-inc/exql/blob/main/changelogs/v2.0.md#exqlquery-package) for more details.

```go
{{.QueryBuilder}}
```

## License

MIT License / Copyright (c) LoiLo inc.

