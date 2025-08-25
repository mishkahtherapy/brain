package ports

import "database/sql"

type SQLExec interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
}

type SQLTx interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)

	Commit() error
	Rollback() error
}

type SQLDatabase interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
	Begin() (SQLTx, error)
	Close() error
}

// TODO: Apply transactions to repos
type TransactionPort interface {
	Begin() (SQLTx, error)
	Commit(tx SQLTx) error
	Rollback(tx SQLTx) error
}
