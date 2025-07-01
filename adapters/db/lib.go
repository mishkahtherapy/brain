package db

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/mishkahtherapy/brain/core/ports"
)

type Database struct {
	db *sql.DB
}

type DatabaseConfig struct {
	// Host     string
	// Port     int
	// User     string
	// Password string
	DBFilename string
	SchemaFile string
}

func NewDatabase(config DatabaseConfig) ports.SQLDatabase {
	if config.SchemaFile == "" {
		panic("schema file is required")
	}
	if config.DBFilename == "" {
		panic("db filename is required")
	}

	db, err := connectDB(config)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		panic(err)
	}
	return &Database{db: db}
}

func (d *Database) Query(query string, args ...any) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

func (d *Database) QueryRow(query string, args ...any) *sql.Row {
	return d.db.QueryRow(query, args...)
}

func (d *Database) Exec(query string, args ...any) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

func (d *Database) Begin() (ports.SQLTx, error) {
	return d.db.Begin()
}

func (d *Database) Close() error {
	return d.db.Close()
}

func connectDB(config DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("sqlite", config.DBFilename)
	if err != nil {
		return nil, err
	}

	// Initialize the database
	db.Exec(`PRAGMA foreign_keys = ON`)

	// Check if the database is has no schema tables
	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name='specializations'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		// Database has schema tables, so we don't need to load the schema
		return db, nil
	}

	// Load schema
	schema, err := os.ReadFile(config.SchemaFile)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return nil, err
	}

	return db, nil
}
