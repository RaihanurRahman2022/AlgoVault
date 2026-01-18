package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Password  string    `json:"-"`    // Never return password in JSON
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
	ID        string    `json:"id"`
	ProblemID string    `json:"problemId"`
	Language  string    `json:"language"` // cpp, go, python, java, javascript
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Problem represents a coding problem
type Problem struct {
	ID           string     `json:"id"`
	PatternID    string     `json:"patternId"`
	Title        string     `json:"title"`
	Difficulty   string     `json:"difficulty"`   // Easy, Medium, Hard
	Description  string     `json:"description"`  // Markdown
	Input        string     `json:"input"`        // Markdown
	Output       string     `json:"output"`       // Markdown
	Constraints  string     `json:"constraints"`  // Markdown
	SampleInput  string     `json:"sampleInput"`  // Markdown
	SampleOutput string     `json:"sampleOutput"` // Markdown
	Explanation  string     `json:"explanation"`  // Markdown
	Notes        string     `json:"notes"`        // Markdown
	Solutions    []Solution `json:"solutions"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// LearningTopic represents a learning category (e.g., LLD, HLD)
type LearningTopic struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Icon        string    `json:"icon"`
	Description string    `json:"description"`
	Slug        string    `json:"slug"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// LearningResource represents a specific resource within a topic
type LearningResource struct {
	ID         string    `json:"id"`
	TopicID    string    `json:"topicId"`
	Title      string    `json:"title"`
	Content    string    `json:"content"` // Markdown
	Type       string    `json:"type"`    // article, video, link
	URL        string    `json:"url"`
	OrderIndex int       `json:"orderIndex"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// RoadmapItem represents a step in a roadmap
type RoadmapItem struct {
	ID          string    `json:"id"`
	TopicID     string    `json:"topicId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	OrderIndex  int       `json:"orderIndex"`
	Status      string    `json:"status"` // todo, in-progress, completed
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Database represents the database connection and operations
type Database struct {
	DB         *sql.DB
	IsPostgres bool
}

// NewDatabase creates a new database connection
// Uses PostgreSQL if DATABASE_URL is set, otherwise uses SQLite
func NewDatabase(dbPath string) (*Database, error) {
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

	if err := database.InitSchema(isPostgres); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %v", err)
	}

	// Run migrations for existing databases
	if err := database.migrate(isPostgres); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	// Create demo user if it doesn't exist
	if err := database.createDemoUser(); err != nil {
		return nil, fmt.Errorf("failed to create demo user: %v", err)
	}

	// Seed initial learning data
	if err := database.seedLearningData(); err != nil {
		log.Printf("Warning: failed to seed learning data: %v", err)
	}

	return database, nil
}

// InitSchema creates all necessary tables
// Supports both SQLite and PostgreSQL
func (d *Database) InitSchema(isPostgres bool) error {
	var timestampType string
	if isPostgres {
		timestampType = "TIMESTAMP DEFAULT CURRENT_TIMESTAMP"
	} else {
		timestampType = "DATETIME DEFAULT CURRENT_TIMESTAMP"
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'admin',
			created_at ` + timestampType + `
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			icon TEXT NOT NULL,
			description TEXT NOT NULL,
			created_at ` + timestampType + `,
			updated_at ` + timestampType + `
		)`,
		`CREATE TABLE IF NOT EXISTS patterns (
			id TEXT PRIMARY KEY,
			category_id TEXT NOT NULL,
			name TEXT NOT NULL,
			icon TEXT NOT NULL,
			description TEXT NOT NULL,
			theory TEXT DEFAULT '',
			created_at ` + timestampType + `,
			updated_at ` + timestampType + `,
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
			created_at ` + timestampType + `,
			updated_at ` + timestampType + `,
			FOREIGN KEY (pattern_id) REFERENCES patterns(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS solutions (
			id TEXT PRIMARY KEY,
			problem_id TEXT NOT NULL,
			language TEXT NOT NULL,
			code TEXT NOT NULL,
			created_at ` + timestampType + `,
			updated_at ` + timestampType + `,
			FOREIGN KEY (problem_id) REFERENCES problems(id) ON DELETE CASCADE,
			UNIQUE(problem_id, language)
		)`,
		`CREATE TABLE IF NOT EXISTS learning_topics (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			icon TEXT NOT NULL,
			description TEXT NOT NULL,
			slug TEXT UNIQUE NOT NULL,
			created_at ` + timestampType + `,
			updated_at ` + timestampType + `
		)`,
		`CREATE TABLE IF NOT EXISTS learning_resources (
			id TEXT PRIMARY KEY,
			topic_id TEXT NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			type TEXT NOT NULL,
			url TEXT,
			order_index INTEGER DEFAULT 0,
			created_at ` + timestampType + `,
			updated_at ` + timestampType + `,
			FOREIGN KEY (topic_id) REFERENCES learning_topics(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS roadmap_items (
			id TEXT PRIMARY KEY,
			topic_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			order_index INTEGER DEFAULT 0,
			status TEXT DEFAULT 'todo',
			created_at ` + timestampType + `,
			updated_at ` + timestampType + `,
			FOREIGN KEY (topic_id) REFERENCES learning_topics(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_patterns_category_id ON patterns(category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_problems_pattern_id ON problems(pattern_id)`,
		`CREATE INDEX IF NOT EXISTS idx_solutions_problem_id ON solutions(problem_id)`,
		`CREATE INDEX IF NOT EXISTS idx_learning_resources_topic_id ON learning_resources(topic_id)`,
		`CREATE INDEX IF NOT EXISTS idx_roadmap_items_topic_id ON roadmap_items(topic_id)`,
	}

	for _, query := range queries {
		if _, err := d.DB.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

// migrate runs database migrations for existing databases
func (d *Database) migrate(isPostgres bool) error {
	// Check if role column exists
	var columnExists bool
	var err error

	if isPostgres {
		// PostgreSQL: Check if column exists
		err = d.DB.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'users' AND column_name = 'role'
			)
		`).Scan(&columnExists)
	} else {
		// SQLite: Use PRAGMA
		rows, err := d.DB.Query("PRAGMA table_info(users)")
		if err != nil {
			return err
		}
		defer rows.Close()

		columnExists = false
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
	}

	if err != nil {
		return err
	}

	if !columnExists {
		// Column doesn't exist, add it
		_, err = d.DB.Exec(`ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'admin'`)
		if err != nil {
			return err
		}

		// Update any existing NULL values to 'admin'
		_, err = d.DB.Exec(`UPDATE users SET role = 'admin' WHERE role IS NULL`)
		if err != nil {
			return err
		}
	}

	// Migrate patterns table to add theory column
	var theoryColumnExists bool
	if isPostgres {
		err = d.DB.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'patterns' AND column_name = 'theory'
			)
		`).Scan(&theoryColumnExists)
	} else {
		patternRows, err := d.DB.Query("PRAGMA table_info(patterns)")
		if err != nil {
			return err
		}
		defer patternRows.Close()

		theoryColumnExists = false
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
	}

	if err != nil {
		return err
	}

	if !theoryColumnExists {
		_, err = d.DB.Exec(`ALTER TABLE patterns ADD COLUMN theory TEXT DEFAULT ''`)
		if err != nil {
			return err
		}
	}

	return nil
}

// convertPlaceholders converts SQLite placeholders (?) to PostgreSQL placeholders ($1, $2, ...)
func (d *Database) convertPlaceholders(query string) string {
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

// createDemoUser creates a demo user with read-only access
func (d *Database) createDemoUser() error {
	// Check if demo user already exists
	var existingID string
	query := d.convertPlaceholders("SELECT id FROM users WHERE email = ?")
	err := d.DB.QueryRow(query, "demo@algovault.com").Scan(&existingID)
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
	query = d.convertPlaceholders("INSERT INTO users (id, email, name, password, role) VALUES (?, ?, ?, ?, ?)")
	_, err = d.DB.Exec(query, demoUserID, "demo@algovault.com", "Demo User", string(hashedPassword), "demo")
	return err
}

// seedLearningData populates initial learning topics
func (d *Database) seedLearningData() error {
	topics := []LearningTopic{
		{ID: "topic-lld", Name: "Low Level Design", Icon: "Layout", Description: "Object-oriented design, design patterns, and SOLID principles.", Slug: "lld"},
		{ID: "topic-hld", Name: "High Level Design", Icon: "Server", Description: "System architecture, scalability, and distributed systems.", Slug: "hld"},
		{ID: "topic-docker", Name: "Docker", Icon: "Box", Description: "Containerization, images, and orchestration basics.", Slug: "docker"},
		{ID: "topic-k8s", Name: "Kubernetes", Icon: "Cloud", Description: "Container orchestration at scale.", Slug: "k8s"},
		{ID: "topic-golang", Name: "Golang", Icon: "Code", Description: "Go programming language, concurrency, and best practices.", Slug: "golang"},
		{ID: "topic-behavioral", Name: "Behavioral", Icon: "Users", Description: "Soft skills and interview preparation.", Slug: "behavioral"},
		{ID: "topic-linux", Name: "Linux", Icon: "Terminal", Description: "Linux commands, shell scripting, and system administration.", Slug: "linux"},
	}

	for _, t := range topics {
		// Check if topic exists
		var exists bool
		query := d.convertPlaceholders("SELECT EXISTS(SELECT 1 FROM learning_topics WHERE slug = ?)")
		err := d.DB.QueryRow(query, t.Slug).Scan(&exists)
		if err != nil {
			return err
		}

		if !exists {
			insertQuery := d.convertPlaceholders("INSERT INTO learning_topics (id, name, icon, description, slug, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)")
			_, err = d.DB.Exec(insertQuery, t.ID, t.Name, t.Icon, t.Description, t.Slug, time.Now(), time.Now())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.DB.Close()
}
