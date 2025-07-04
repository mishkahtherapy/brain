package testutils

import (
	"os"
	"testing"

	"github.com/mishkahtherapy/brain/adapters/db"
	"github.com/mishkahtherapy/brain/core/ports"
)

// SetupTestDB creates a test database with schema and returns cleanup function
// Uses a temporary file database for more realistic testing
func SetupTestDB(t *testing.T) (ports.SQLDatabase, func()) {
	// Create temporary database file
	tmpfile, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	dbFilename := tmpfile.Name()

	database := db.NewDatabase(db.DatabaseConfig{
		DBFilename: dbFilename,
		SchemaFile: "../../../schema.sql",
	})

	// Return cleanup function
	cleanup := func() {
		database.Close()
		os.Remove(dbFilename)
	}

	return database, cleanup
}

// SetupInMemoryTestDB creates an in-memory test database
// Faster than file-based but less realistic
func SetupInMemoryTestDB(t *testing.T) (ports.SQLDatabase, func()) {
	dbFilename := ":memory:"

	database := db.NewDatabase(db.DatabaseConfig{
		DBFilename: dbFilename,
		SchemaFile: "../../../schema.sql",
	})

	// Return cleanup function
	cleanup := func() {
		database.Close()
	}

	return database, cleanup
}
