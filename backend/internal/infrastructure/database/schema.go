package database

import "database/sql"

// InitSchema creates all necessary tables
// Supports both SQLite and PostgreSQL
func (d *Database) InitSchema() error {
	var timestampType string
	if d.IsPostgres {
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

// Migrate runs database migrations for existing databases
func (d *Database) Migrate() error {
	// Check if role column exists
	var columnExists bool
	var err error

	if d.IsPostgres {
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
	if d.IsPostgres {
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
