package main

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "github.com/mattn/go-sqlite3"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Password  string    `json:"-"` // Never return password in JSON
	Role      string    `json:"role"` // "admin" or "demo"
	CreatedAt time.Time `json:"createdAt"`
}

// Category represents a problem category
type Category struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Icon         string    `json:"icon"`
	Description  string    `json:"description"`
	PatternCount int       `json:"patternCount"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Pattern represents a problem pattern under a category
type Pattern struct {
	ID           string    `json:"id"`
	CategoryID   string    `json:"categoryId"`
	Name         string    `json:"name"`
	Icon         string    `json:"icon"`
	Description  string    `json:"description"`
	Theory       string    `json:"theory"` // Markdown content
	ProblemCount int       `json:"problemCount"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Solution represents a solution in a specific language
type Solution struct {
	ID         string `json:"id"`
	ProblemID  string `json:"problemId"`
	Language   string `json:"language"` // cpp, go, python, java, javascript
	Code       string `json:"code"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// Problem represents a coding problem
type Problem struct {
	ID           string     `json:"id"`
	PatternID    string     `json:"patternId"`
	Title        string     `json:"title"`
	Difficulty   string     `json:"difficulty"` // Easy, Medium, Hard
	Description  string     `json:"description"` // Markdown
	Input        string     `json:"input"`       // Markdown
	Output       string     `json:"output"`     // Markdown
	Constraints  string     `json:"constraints"` // Markdown
	SampleInput  string     `json:"sampleInput"` // Markdown
	SampleOutput string     `json:"sampleOutput"` // Markdown
	Explanation  string     `json:"explanation"`  // Markdown
	Notes        string     `json:"notes"`        // Markdown
	Solutions    []Solution `json:"solutions"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// Database represents the database connection and operations
type Database struct {
	DB *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database := &Database{DB: db}
	if err := database.InitSchema(); err != nil {
		return nil, err
	}

	// Run migrations for existing databases
	if err := database.migrate(); err != nil {
		return nil, err
	}

	// Create demo user if it doesn't exist
	if err := database.createDemoUser(); err != nil {
		return nil, err
	}

	return database, nil
}

// InitSchema creates all necessary tables
func (d *Database) InitSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'admin',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			icon TEXT NOT NULL,
			description TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS patterns (
			id TEXT PRIMARY KEY,
			category_id TEXT NOT NULL,
			name TEXT NOT NULL,
			icon TEXT NOT NULL,
			description TEXT NOT NULL,
			theory TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS problems (
			id TEXT PRIMARY KEY,
			pattern_id TEXT NOT NULL,
			title TEXT NOT NULL,
			difficulty TEXT NOT NULL,
			description TEXT NOT NULL,
			input TEXT NOT NULL,
			output TEXT NOT NULL,
			constraints TEXT NOT NULL,
			sample_input TEXT NOT NULL,
			sample_output TEXT NOT NULL,
			explanation TEXT NOT NULL,
			notes TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (pattern_id) REFERENCES patterns(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS solutions (
			id TEXT PRIMARY KEY,
			problem_id TEXT NOT NULL,
			language TEXT NOT NULL,
			code TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (problem_id) REFERENCES problems(id) ON DELETE CASCADE,
			UNIQUE(problem_id, language)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_patterns_category_id ON patterns(category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_problems_pattern_id ON problems(pattern_id)`,
		`CREATE INDEX IF NOT EXISTS idx_solutions_problem_id ON solutions(problem_id)`,
	}

	for _, query := range queries {
		if _, err := d.DB.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

// migrate runs database migrations for existing databases
func (d *Database) migrate() error {
	// Check if role column exists by querying the table schema
	rows, err := d.DB.Query("PRAGMA table_info(users)")
	if err != nil {
		return err
	}
	defer rows.Close()
	
	columnExists := false
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return err
		}
		
		if name == "role" {
			columnExists = true
			break
		}
	}
	
	if !columnExists {
		// Column doesn't exist, add it
		// SQLite allows adding a column with DEFAULT even if table has rows
		_, err = d.DB.Exec(`
			ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'admin'
		`)
		if err != nil {
			return err
		}

		// Update any existing NULL values to 'admin' (shouldn't be needed with DEFAULT, but just in case)
		_, err = d.DB.Exec(`
			UPDATE users SET role = 'admin' WHERE role IS NULL
		`)
		if err != nil {
			return err
		}
	}

	// Migrate patterns table to add theory column
	patternRows, err := d.DB.Query("PRAGMA table_info(patterns)")
	if err != nil {
		return err
	}
	defer patternRows.Close()

	theoryColumnExists := false
	for patternRows.Next() {
		var cid int
		var name, dataType string
		var notNull int
		var defaultValue sql.NullString
		var pk int

		if err := patternRows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return err
		}

		if name == "theory" {
			theoryColumnExists = true
			break
		}
	}

	if !theoryColumnExists {
		_, err = d.DB.Exec(`
			ALTER TABLE patterns ADD COLUMN theory TEXT DEFAULT ''
		`)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// createDemoUser creates a demo user with read-only access
func (d *Database) createDemoUser() error {
	// Check if demo user already exists
	var existingID string
	err := d.DB.QueryRow("SELECT id FROM users WHERE email = ?", "demo@algovault.com").Scan(&existingID)
	if err == nil {
		// Demo user already exists
		return nil
	}
	if err != sql.ErrNoRows {
		return err
	}

	// Create demo user with password "demo123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("demo123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	demoUserID := "demo-user-001"
	_, err = d.DB.Exec(
		"INSERT INTO users (id, email, name, password, role) VALUES (?, ?, ?, ?, ?)",
		demoUserID, "demo@algovault.com", "Demo User", string(hashedPassword), "demo",
	)
	return err
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.DB.Close()
}
