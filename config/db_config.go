package config

import (
	"path/filepath"

	"github.com/mishkahtherapy/brain/adapters/db"
)

func GetDBConfig() db.DatabaseConfig {
	dbRootPath := MustGetEnv("BRAIN_DATABASE_PATH")
	// Join path with brain.db
	dbPath := filepath.Join(dbRootPath, "brain.db")
	schemaPath := filepath.Join(dbRootPath, "schema.sql")
	return db.DatabaseConfig{
		DBFilename: dbPath,
		SchemaFile: schemaPath,
	}
}
