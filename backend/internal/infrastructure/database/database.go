package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Database represents the database connection and operations
type Database struct {
	DB         *sql.DB
	IsPostgres bool
}

// New creates a new database connection
// Uses PostgreSQL if DATABASE_URL is set, otherwise uses SQLite
func New(dbPath string) (*Database, error) {
	var db *sql.DB
	var err error
	var isPostgres bool

	// Check if DATABASE_URL is set (PostgreSQL)
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		// Use PostgreSQL
		db, err = sql.Open("postgres", databaseURL)
		isPostgres = true
		if err != nil {
			return nil, fmt.Errorf("failed to open PostgreSQL connection: %v", err)
		}
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
	} else {
		// Use SQLite with WAL mode and busy timeout
		db, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
		isPostgres = false
		if err != nil {
			return nil, fmt.Errorf("failed to open SQLite connection: %v", err)
		}
		// SQLite doesn't handle multiple connections well
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	database := &Database{DB: db, IsPostgres: isPostgres}

	if err := database.InitSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %v", err)
	}

	// Run migrations for existing databases
	if err := database.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.DB.Close()
}

// ConvertPlaceholders converts SQLite placeholders (?) to PostgreSQL placeholders ($1, $2, ...)
func (d *Database) ConvertPlaceholders(query string) string {
	if !d.IsPostgres {
		return query // SQLite uses ?
	}
	// Convert ? to $1, $2, etc. for PostgreSQL
	placeholderNum := 1
	result := ""
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			result += "$" + fmt.Sprintf("%d", placeholderNum)
			placeholderNum++
		} else {
			result += string(query[i])
		}
	}
	return result
}
