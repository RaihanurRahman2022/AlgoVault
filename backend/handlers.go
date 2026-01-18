package main

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type Handlers struct {
	DB        *Database
	JWTSecret string
	AIAPIKey  string
}

// Auth handlers

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type ThitaProblemResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Difficulty  string `json:"difficulty"`
	CodeStubs   []struct {
		Language string `json:"language"`
		CodeStub string `json:"code_stub"`
	} `json:"code_stubs"`
	Hints     []string `json:"hints"`
	TestCases []struct {
		InputData      string `json:"input_data"`
		ExpectedOutput string `json:"expected_output"`
		Explanation    string `json:"explanation"`
	} `json:"test_cases"`
}

type ThitaBulkResponse struct {
	Categories []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Patterns    []struct {
			ID              int    `json:"id"`
			Name            string `json:"name"`
			Description     string `json:"description"`
			MatchedProblems []struct {
				ID         string `json:"id"`
				Title      string `json:"title"`
				Difficulty string `json:"difficulty"`
			} `json:"matched_problems"`
		} `json:"patterns"`
	} `json:"categories"`
}

// Login handles user authentication
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	var user User
	var role sql.NullString
	// Normalize email for comparison (trim and lowercase)
	normalizedEmail := strings.TrimSpace(strings.ToLower(req.Email))
	trimmedEmail := strings.TrimSpace(req.Email)

	// Use COALESCE to handle NULL role values (default to 'admin')
	// Try multiple query strategies to find the user
	var err error

	// Strategy 1: Exact match (case-sensitive, trimmed)
	query1 := h.DB.convertPlaceholders("SELECT id, email, name, password, COALESCE(role, 'admin') as role FROM users WHERE email = ?")
	err = h.DB.DB.QueryRow(query1, trimmedEmail).Scan(&user.ID, &user.Email, &user.Name, &user.Password, &role)

	// Strategy 2: Case-insensitive match
	if err == sql.ErrNoRows {
		query2 := h.DB.convertPlaceholders("SELECT id, email, name, password, COALESCE(role, 'admin') as role FROM users WHERE LOWER(email) = ?")
		err = h.DB.DB.QueryRow(query2, normalizedEmail).Scan(&user.ID, &user.Email, &user.Name, &user.Password, &role)
	}

	// Strategy 3: Query all users and find match manually (fallback for edge cases)
	if err == sql.ErrNoRows {
		rows, queryErr := h.DB.DB.Query("SELECT id, email, name, password, COALESCE(role, 'admin') as role FROM users")
		if queryErr == nil {
			defer rows.Close()
			for rows.Next() {
				var tempUser User
				var tempRole sql.NullString
				if scanErr := rows.Scan(&tempUser.ID, &tempUser.Email, &tempUser.Name, &tempUser.Password, &tempRole); scanErr == nil {
					// Compare normalized emails
					if strings.TrimSpace(strings.ToLower(tempUser.Email)) == normalizedEmail {
						user = tempUser
						role = tempRole
						err = nil
						break
					}
				}
			}
		}
	}

	if err == sql.ErrNoRows {
		// Return detailed error message to help debug
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("User not found. Email tried: '%s'. Please check your email or register first.", req.Email))
		return
	}
	if err != nil {
		// Return the actual database error to help debug
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Database error: %v", err))
		return
	}

	// Handle role
	if role.Valid && role.String != "" {
		user.Role = role.String
	} else {
		user.Role = "admin"
		// Update the user in database to have a role if it was NULL
		updateQuery := h.DB.convertPlaceholders("UPDATE users SET role = 'admin' WHERE id = ? AND (role IS NULL OR role = '')")
		h.DB.DB.Exec(updateQuery, user.ID)
	}

	// Compare password - check if password hash is valid first
	if len(user.Password) == 0 {
		respondWithError(w, http.StatusInternalServerError, "User account error: password hash is empty")
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		// Return more specific error (but don't reveal too much for security)
		respondWithError(w, http.StatusUnauthorized, "Invalid password. Please check your password and try again.")
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": user.ID,
		"exp":    time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	})

	tokenString, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"token": tokenString,
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		},
	})
}

// Register handles user registration
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Email, password, and name are required")
		return
	}

	// Check if user already exists
	var existingID string
	query := h.DB.convertPlaceholders("SELECT id FROM users WHERE email = ? OR LOWER(email) = LOWER(?)")
	err := h.DB.DB.QueryRow(query, req.Email, req.Email).Scan(&existingID)
	if err == nil {
		respondWithError(w, http.StatusConflict, "User with this email already exists")
		return
	}
	if err != sql.ErrNoRows {
		respondWithError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password")
		return
	}

	// Generate user ID
	userID := generateID()

	// Insert user (default role is 'admin' as per schema)
	query = h.DB.convertPlaceholders("INSERT INTO users (id, email, name, password, role) VALUES (?, ?, ?, ?, ?)")
	_, err = h.DB.DB.Exec(
		query,
		userID, req.Email, req.Name, string(hashedPassword), "admin",
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating user: "+err.Error())
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"exp":    time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	})

	tokenString, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"token": tokenString,
		"user": map[string]interface{}{
			"id":    userID,
			"email": req.Email,
			"name":  req.Name,
			"role":  "admin",
		},
	})
}

// Category handlers

func (h *Handlers) GetCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.DB.Query(`
		SELECT c.id, c.name, c.icon, c.description, c.created_at, c.updated_at,
		       COUNT(DISTINCT p.id) as pattern_count
		FROM categories c
		LEFT JOIN patterns p ON p.category_id = c.id
		GROUP BY c.id, c.name, c.icon, c.description, c.created_at, c.updated_at
		ORDER BY c.created_at ASC
	`)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var cat Category
		err := rows.Scan(&cat.ID, &cat.Name, &cat.Icon, &cat.Description, &cat.CreatedAt, &cat.UpdatedAt, &cat.PatternCount)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning category")
			return
		}
		categories = append(categories, cat)
	}

	respondWithJSON(w, http.StatusOK, categories)
}

func (h *Handlers) CreateCategory(w http.ResponseWriter, r *http.Request) {
	if isDemoUser(r) {
		respondWithError(w, http.StatusForbidden, "Demo users cannot create categories")
		return
	}
	var cat Category
	if err := json.NewDecoder(r.Body).Decode(&cat); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cat.ID = generateID()
	cat.PatternCount = 0
	cat.CreatedAt = time.Now()
	cat.UpdatedAt = time.Now()

	query := h.DB.convertPlaceholders("INSERT INTO categories (id, name, icon, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)")
	_, err := h.DB.DB.Exec(query, cat.ID, cat.Name, cat.Icon, cat.Description, cat.CreatedAt, cat.UpdatedAt)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating category")
		return
	}

	respondWithJSON(w, http.StatusCreated, cat)
}

func (h *Handlers) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	if isDemoUser(r) {
		respondWithError(w, http.StatusForbidden, "Demo users cannot update categories")
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]

	var cat Category
	if err := json.NewDecoder(r.Body).Decode(&cat); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cat.UpdatedAt = time.Now()
	query := h.DB.convertPlaceholders("UPDATE categories SET name = ?, icon = ?, description = ?, updated_at = ? WHERE id = ?")
	_, err := h.DB.DB.Exec(query, cat.Name, cat.Icon, cat.Description, cat.UpdatedAt, id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating category")
		return
	}

	cat.ID = id
	respondWithJSON(w, http.StatusOK, cat)
}

func (h *Handlers) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	if isDemoUser(r) {
		respondWithError(w, http.StatusForbidden, "Demo users cannot delete categories")
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]

	query := h.DB.convertPlaceholders("DELETE FROM categories WHERE id = ?")
	_, err := h.DB.DB.Exec(query, id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting category")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Category deleted"})
}

// Pattern handlers

func (h *Handlers) GetPatterns(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID := vars["categoryId"]

	query := h.DB.convertPlaceholders(`
		SELECT p.id, p.category_id, p.name, p.icon, p.description, COALESCE(p.theory, '') as theory, p.created_at, p.updated_at,
		       COUNT(DISTINCT pr.id) as problem_count
		FROM patterns p
		LEFT JOIN problems pr ON pr.pattern_id = p.id
		WHERE p.category_id = ?
		GROUP BY p.id, p.category_id, p.name, p.icon, p.description, p.theory, p.created_at, p.updated_at
		ORDER BY p.created_at ASC
	`)
	rows, err := h.DB.DB.Query(query, categoryID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Database error: %v", err))
		return
	}
	defer rows.Close()

	var patterns []Pattern
	for rows.Next() {
		var pat Pattern
		err := rows.Scan(&pat.ID, &pat.CategoryID, &pat.Name, &pat.Icon, &pat.Description, &pat.Theory, &pat.CreatedAt, &pat.UpdatedAt, &pat.ProblemCount)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning pattern")
			return
		}
		patterns = append(patterns, pat)
	}

	respondWithJSON(w, http.StatusOK, patterns)
}

func (h *Handlers) CreatePattern(w http.ResponseWriter, r *http.Request) {
	if isDemoUser(r) {
		respondWithError(w, http.StatusForbidden, "Demo users cannot create patterns")
		return
	}
	vars := mux.Vars(r)
	categoryID := vars["categoryId"]

	var pat Pattern
	if err := json.NewDecoder(r.Body).Decode(&pat); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	pat.ID = generateID()
	pat.CategoryID = categoryID
	pat.ProblemCount = 0
	pat.CreatedAt = time.Now()
	pat.UpdatedAt = time.Now()

	query := h.DB.convertPlaceholders("INSERT INTO patterns (id, category_id, name, icon, description, theory, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	_, err := h.DB.DB.Exec(query, pat.ID, pat.CategoryID, pat.Name, pat.Icon, pat.Description, pat.Theory, pat.CreatedAt, pat.UpdatedAt)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating pattern")
		return
	}

	respondWithJSON(w, http.StatusCreated, pat)
}

func (h *Handlers) UpdatePattern(w http.ResponseWriter, r *http.Request) {
	if isDemoUser(r) {
		respondWithError(w, http.StatusForbidden, "Demo users cannot update patterns")
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]

	var pat Pattern
	if err := json.NewDecoder(r.Body).Decode(&pat); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	pat.UpdatedAt = time.Now()
	query := h.DB.convertPlaceholders("UPDATE patterns SET name = ?, icon = ?, description = ?, theory = ?, updated_at = ? WHERE id = ?")
	_, err := h.DB.DB.Exec(query, pat.Name, pat.Icon, pat.Description, pat.Theory, pat.UpdatedAt, id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating pattern")
		return
	}

	pat.ID = id
	respondWithJSON(w, http.StatusOK, pat)
}

func (h *Handlers) DeletePattern(w http.ResponseWriter, r *http.Request) {
	if isDemoUser(r) {
		respondWithError(w, http.StatusForbidden, "Demo users cannot delete patterns")
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]

	query := h.DB.convertPlaceholders("DELETE FROM patterns WHERE id = ?")
	_, err := h.DB.DB.Exec(query, id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting pattern")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Pattern deleted"})
}

// UpdatePatternTheory updates only the theory field of a pattern
func (h *Handlers) UpdatePatternTheory(w http.ResponseWriter, r *http.Request) {
	if isDemoUser(r) {
		respondWithError(w, http.StatusForbidden, "Demo users cannot update pattern theory")
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]

	var req struct {
		Theory string `json:"theory"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	query := h.DB.convertPlaceholders("UPDATE patterns SET theory = ?, updated_at = ? WHERE id = ?")
	_, err := h.DB.DB.Exec(query, req.Theory, time.Now(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating pattern theory")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Pattern theory updated"})
}

// Problem handlers

func (h *Handlers) GetProblems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	patternID := vars["patternId"]

	query := h.DB.convertPlaceholders(`
		SELECT id, pattern_id, title, difficulty, description, input, output, 
		       constraints, sample_input, sample_output, explanation, notes,
		       created_at, updated_at
		FROM problems
		WHERE pattern_id = ?
		ORDER BY created_at ASC
	`)
	rows, err := h.DB.DB.Query(query, patternID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var problems []Problem
	for rows.Next() {
		var prob Problem
		err := rows.Scan(
			&prob.ID, &prob.PatternID, &prob.Title, &prob.Difficulty,
			&prob.Description, &prob.Input, &prob.Output, &prob.Constraints,
			&prob.SampleInput, &prob.SampleOutput, &prob.Explanation, &prob.Notes,
			&prob.CreatedAt, &prob.UpdatedAt,
		)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning problem")
			return
		}

		// Load solutions for this problem
		solutions, err := h.getSolutions(prob.ID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error loading solutions")
			return
		}
		prob.Solutions = solutions

		problems = append(problems, prob)
	}

	respondWithJSON(w, http.StatusOK, problems)
}

func (h *Handlers) GetProblem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var prob Problem
	query := h.DB.convertPlaceholders(`
		SELECT id, pattern_id, title, difficulty, description, input, output,
		       constraints, sample_input, sample_output, explanation, notes,
		       created_at, updated_at
		FROM problems
		WHERE id = ?
	`)
	err := h.DB.DB.QueryRow(query, id).Scan(
		&prob.ID, &prob.PatternID, &prob.Title, &prob.Difficulty,
		&prob.Description, &prob.Input, &prob.Output, &prob.Constraints,
		&prob.SampleInput, &prob.SampleOutput, &prob.Explanation, &prob.Notes,
		&prob.CreatedAt, &prob.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Problem not found")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Load solutions
	solutions, err := h.getSolutions(prob.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error loading solutions")
		return
	}
	prob.Solutions = solutions

	respondWithJSON(w, http.StatusOK, prob)
}

func (h *Handlers) CreateProblem(w http.ResponseWriter, r *http.Request) {
	if isDemoUser(r) {
		respondWithError(w, http.StatusForbidden, "Demo users cannot create problems")
		return
	}
	vars := mux.Vars(r)
	patternID := vars["patternId"]

	var prob Problem
	if err := json.NewDecoder(r.Body).Decode(&prob); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	prob.ID = generateID()
	prob.PatternID = patternID
	prob.CreatedAt = time.Now()
	prob.UpdatedAt = time.Now()

	query := h.DB.convertPlaceholders(`INSERT INTO problems (id, pattern_id, title, difficulty, description, input, output, constraints, sample_input, sample_output, explanation, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	_, err := h.DB.DB.Exec(query, prob.ID, prob.PatternID, prob.Title, prob.Difficulty,
		prob.Description, prob.Input, prob.Output, prob.Constraints,
		prob.SampleInput, prob.SampleOutput, prob.Explanation, prob.Notes,
		prob.CreatedAt, prob.UpdatedAt)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating problem")
		return
	}

	// Save solutions
	for _, sol := range prob.Solutions {
		sol.ID = generateID()
		sol.ProblemID = prob.ID
		sol.CreatedAt = time.Now()
		sol.UpdatedAt = time.Now()
		solQuery := h.DB.convertPlaceholders(`INSERT INTO solutions (id, problem_id, language, code, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`)
		_, err := h.DB.DB.Exec(solQuery, sol.ID, sol.ProblemID, sol.Language, sol.Code, sol.CreatedAt, sol.UpdatedAt)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error saving solution")
			return
		}
	}

	// Reload solutions
	solutions, err := h.getSolutions(prob.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error loading solutions")
		return
	}
	prob.Solutions = solutions

	respondWithJSON(w, http.StatusCreated, prob)
}

func (h *Handlers) UpdateProblem(w http.ResponseWriter, r *http.Request) {
	if isDemoUser(r) {
		respondWithError(w, http.StatusForbidden, "Demo users cannot update problems")
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]

	var prob Problem
	if err := json.NewDecoder(r.Body).Decode(&prob); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	prob.UpdatedAt = time.Now()
	query := h.DB.convertPlaceholders(`UPDATE problems SET title = ?, difficulty = ?, description = ?, input = ?, output = ?, constraints = ?, sample_input = ?, sample_output = ?, explanation = ?, notes = ?, updated_at = ? WHERE id = ?`)
	_, err := h.DB.DB.Exec(query, prob.Title, prob.Difficulty, prob.Description, prob.Input, prob.Output,
		prob.Constraints, prob.SampleInput, prob.SampleOutput, prob.Explanation,
		prob.Notes, prob.UpdatedAt, id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating problem")
		return
	}

	// Update solutions - use UPSERT based on (problem_id, language) unique constraint
	for _, sol := range prob.Solutions {
		sol.ProblemID = id
		sol.UpdatedAt = time.Now()

		// Check if solution exists for this language
		var existingID string
		checkQuery := h.DB.convertPlaceholders(`SELECT id FROM solutions WHERE problem_id = ? AND language = ?`)
		err := h.DB.DB.QueryRow(checkQuery, id, sol.Language).Scan(&existingID)

		if err == sql.ErrNoRows {
			// New solution - generate ID
			sol.ID = generateID()
			sol.CreatedAt = time.Now()
			insertQuery := h.DB.convertPlaceholders(`INSERT INTO solutions (id, problem_id, language, code, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`)
			_, err = h.DB.DB.Exec(insertQuery, sol.ID, sol.ProblemID, sol.Language, sol.Code, sol.CreatedAt, sol.UpdatedAt)
		} else if err == nil {
			// Existing solution - update it
			sol.ID = existingID
			updateQuery := h.DB.convertPlaceholders(`UPDATE solutions SET code = ?, updated_at = ? WHERE problem_id = ? AND language = ?`)
			_, err = h.DB.DB.Exec(updateQuery, sol.Code, sol.UpdatedAt, sol.ProblemID, sol.Language)
		}

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error saving solution")
			return
		}
	}

	prob.ID = id
	// Reload solutions
	solutions, err := h.getSolutions(prob.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error loading solutions")
		return
	}
	prob.Solutions = solutions

	respondWithJSON(w, http.StatusOK, prob)
}

func (h *Handlers) DeleteProblem(w http.ResponseWriter, r *http.Request) {
	if isDemoUser(r) {
		respondWithError(w, http.StatusForbidden, "Demo users cannot delete problems")
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]

	query := h.DB.convertPlaceholders("DELETE FROM problems WHERE id = ?")
	_, err := h.DB.DB.Exec(query, id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting problem")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Problem deleted"})
}

// Helper function to get solutions for a problem
func (h *Handlers) getSolutions(problemID string) ([]Solution, error) {
	query := h.DB.convertPlaceholders(`
		SELECT id, problem_id, language, code, created_at, updated_at
		FROM solutions
		WHERE problem_id = ?
	`)
	rows, err := h.DB.DB.Query(query, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var solutions []Solution
	for rows.Next() {
		var sol Solution
		err := rows.Scan(&sol.ID, &sol.ProblemID, &sol.Language, &sol.Code, &sol.CreatedAt, &sol.UpdatedAt)
		if err != nil {
			return nil, err
		}
		solutions = append(solutions, sol)
	}

	return solutions, nil
}

// generateID generates a unique ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// AI Problem Generation

type GenerateProblemRequest struct {
	Query string `json:"query"`
}

type GenerateProblemResponse struct {
	Title        string `json:"title"`
	Difficulty   string `json:"difficulty"`
	Description  string `json:"description"`
	Input        string `json:"input"`
	Output       string `json:"output"`
	Constraints  string `json:"constraints"`
	SampleInput  string `json:"sampleInput"`
	SampleOutput string `json:"sampleOutput"`
	Explanation  string `json:"explanation"`
	Notes        string `json:"notes"`
}

// GenerateProblem uses AI to generate problem details
func (h *Handlers) GenerateProblem(w http.ResponseWriter, r *http.Request) {
	var req GenerateProblemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Query == "" {
		respondWithError(w, http.StatusBadRequest, "Query is required")
		return
	}

	// Prepare prompt for AI
	prompt := fmt.Sprintf(`Generate a complete coding problem description based on the following query: "%s"

Please provide a well-structured coding problem with the following sections in JSON format:
- title: A clear, concise problem title
- difficulty: One of "Easy", "Medium", or "Hard"
- description: A detailed problem description in markdown format explaining what needs to be solved
- input: Description of the input format in markdown
- output: Description of the expected output format in markdown
- constraints: Problem constraints in markdown (e.g., time limits, space limits, value ranges)
- sampleInput: A sample input example
- sampleOutput: The corresponding expected output for the sample input
- explanation: A brief explanation of the sample input/output in markdown

Return ONLY valid JSON with these exact keys. Use markdown formatting for multi-line text fields.`, req.Query)

	// Call OpenRouter API
	openRouterURL := "https://openrouter.ai/api/v1/chat/completions"
	requestBody := map[string]interface{}{
		"model": "openai/gpt-4o-mini",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error preparing request")
		return
	}

	httpReq, err := http.NewRequest("POST", openRouterURL, bytes.NewBuffer(jsonData))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating request")
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+h.AIAPIKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/algovault")
	httpReq.Header.Set("X-Title", "AlgoVault")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error calling AI service: "+err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reading response")
		return
	}

	if resp.StatusCode != http.StatusOK {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("AI service error: %s", string(body)))
		return
	}

	// Parse OpenRouter response
	var openRouterResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error parsing AI response")
		return
	}

	if len(openRouterResp.Choices) == 0 {
		respondWithError(w, http.StatusInternalServerError, "No response from AI")
		return
	}

	// Extract JSON from the response (might be wrapped in markdown code blocks)
	content := openRouterResp.Choices[0].Message.Content

	// Try to extract JSON from markdown code blocks if present
	jsonStart := 0
	jsonEnd := len(content)
	contentBytes := []byte(content)
	if idx := bytes.Index(contentBytes, []byte("```json")); idx != -1 {
		jsonStart = idx + 7
	} else if idx := bytes.Index(contentBytes, []byte("```")); idx != -1 {
		jsonStart = idx + 3
	}
	if idx := bytes.LastIndex(contentBytes, []byte("```")); idx != -1 && idx > jsonStart {
		jsonEnd = idx
	}

	jsonContent := string(bytes.TrimSpace([]byte(content[jsonStart:jsonEnd])))

	// Parse the generated problem
	var generatedProblem GenerateProblemResponse
	if err := json.Unmarshal([]byte(jsonContent), &generatedProblem); err != nil {
		// If JSON parsing fails, try to create a basic structure
		respondWithError(w, http.StatusInternalServerError, "Error parsing AI response: "+err.Error()+". Raw content: "+string(jsonContent[:min(200, len(jsonContent))]))
		return
	}

	// Validate difficulty
	if generatedProblem.Difficulty != "Easy" && generatedProblem.Difficulty != "Medium" && generatedProblem.Difficulty != "Hard" {
		generatedProblem.Difficulty = "Medium" // Default
	}

	respondWithJSON(w, http.StatusOK, generatedProblem)
}

// FetchExternalProblem fetches problem details from Thita.ai API
func (h *Handlers) FetchExternalProblem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	problemID := vars["problemId"]

	if problemID == "" {
		respondWithError(w, http.StatusBadRequest, "Problem ID is required")
		return
	}

	// Clean the problem ID (remove leading slash if present)
	problemID = strings.TrimPrefix(problemID, "/")

	url := fmt.Sprintf("https://api.thita.ai/api/technical-coaching/problems/%s", problemID)

	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "AlgoVault/1.0")

	resp, err := client.Do(req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to connect to external API: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respondWithError(w, resp.StatusCode, fmt.Sprintf("External API returned error status: %d", resp.StatusCode))
		return
	}

	var thitaResp ThitaProblemResponse
	if err := json.NewDecoder(resp.Body).Decode(&thitaResp); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to parse external API response")
		return
	}

	// Helper to clean markdown and decode HTML entities
	cleanMarkdown := func(s string) string {
		s = strings.ReplaceAll(s, "&nbsp;", " ")
		s = strings.ReplaceAll(s, "&lt;", "<")
		s = strings.ReplaceAll(s, "&gt;", ">")
		s = strings.ReplaceAll(s, "&amp;", "&")
		s = strings.ReplaceAll(s, "\u0026nbsp;", " ")
		s = strings.ReplaceAll(s, "\u0026lt;", "<")
		s = strings.ReplaceAll(s, "\u0026gt;", ">")
		s = strings.ReplaceAll(s, "\u0026amp;", "&")
		s = strings.ReplaceAll(s, "  ", " ") // Remove double spaces
		return strings.TrimSpace(s)
	}

	// Map Thita response to our format
	result := GenerateProblemResponse{
		Title:      thitaResp.Title,
		Difficulty: thitaResp.Difficulty,
	}

	// Parse description to extract sections
	desc := thitaResp.Description

	// Try to extract constraints first (usually at the end)
	constraints := ""
	if idx := strings.Index(desc, "**Constraints:**"); idx != -1 {
		constraints = cleanMarkdown(desc[idx+len("**Constraints:**"):])
		desc = desc[:idx]
	}

	// Try to extract Input/Output format if they exist in description

	inputFormat := "See description"
	outputFormat := "See description"

	// Check for explicit format sections
	inputMarkers := []string{"**Input Format**", "**Input format**", "### Input"}
	outputMarkers := []string{"**Output Format**", "**Output format**", "### Output"}

	for _, m := range inputMarkers {
		if idx := strings.Index(desc, m); idx != -1 {
			endIdx := len(desc)
			for _, nextM := range append(outputMarkers, "**Example", "Example 1", "**Constraints") {
				if nIdx := strings.Index(desc[idx+len(m):], nextM); nIdx != -1 {
					if idx+len(m)+nIdx < endIdx {
						endIdx = idx + len(m) + nIdx
					}
				}
			}
			inputFormat = cleanMarkdown(desc[idx+len(m) : endIdx])
			break
		}
	}

	for _, m := range outputMarkers {
		if idx := strings.Index(desc, m); idx != -1 {
			endIdx := len(desc)
			for _, nextM := range []string{"**Example", "Example 1", "**Constraints"} {
				if nIdx := strings.Index(desc[idx+len(m):], nextM); nIdx != -1 {
					if idx+len(m)+nIdx < endIdx {
						endIdx = idx + len(m) + nIdx
					}
				}
			}
			outputFormat = cleanMarkdown(desc[idx+len(m) : endIdx])
			break
		}
	}

	// Split description before examples to keep it clean
	exampleMarkers := []string{"**Example 1:**", "Example 1:", "**Example:**", "Example:", "**Example 1**", "Example 1"}
	for _, marker := range exampleMarkers {
		if idx := strings.Index(desc, marker); idx != -1 {
			desc = desc[:idx]
			break
		}
	}

	result.Description = cleanMarkdown(desc)
	result.Constraints = constraints
	if result.Constraints == "" {
		result.Constraints = "No specific constraints provided."
	}
	result.Input = inputFormat
	result.Output = outputFormat

	// Try to extract sample input/output from test cases
	if len(thitaResp.TestCases) > 0 {
		result.SampleInput = cleanMarkdown(thitaResp.TestCases[0].InputData)
		result.SampleOutput = cleanMarkdown(thitaResp.TestCases[0].ExpectedOutput)
		result.Explanation = cleanMarkdown(thitaResp.TestCases[0].Explanation)
	}

	// Map hints to notes
	if len(thitaResp.Hints) > 0 {
		result.Notes = "### Hints\n"
		for _, hint := range thitaResp.Hints {
			result.Notes += fmt.Sprintf("- %s\n", cleanMarkdown(hint))
		}
	}

	respondWithJSON(w, http.StatusOK, result)
}

// FetchAllExternalData fetches the entire DSA pattern structure from Thita.ai
func (h *Handlers) FetchAllExternalData(w http.ResponseWriter, r *http.Request) {
	if isDemoUser(r) {
		respondWithError(w, http.StatusForbidden, "Demo users cannot fetch bulk data")
		return
	}

	url := "https://api.thita.ai/api/technical-coaching/dsa-pattern-structure?refresh=true"

	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "AlgoVault/1.0")

	resp, err := client.Do(req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to connect to external API: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respondWithError(w, resp.StatusCode, fmt.Sprintf("External API returned error status: %d", resp.StatusCode))
		return
	}

	var thitaResp ThitaBulkResponse
	if err := json.NewDecoder(resp.Body).Decode(&thitaResp); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to parse external API response")
		return
	}

	// Process categories, patterns, and problems
	countCategories := 0
	countPatterns := 0
	countProblems := 0

	for _, tCat := range thitaResp.Categories {
		// 1. Check if category exists or create it
		var catID string
		query := h.DB.convertPlaceholders("SELECT id FROM categories WHERE name = ?")
		err := h.DB.DB.QueryRow(query, tCat.Name).Scan(&catID)

		if err == sql.ErrNoRows {
			catID = generateID()
			insertQuery := h.DB.convertPlaceholders("INSERT INTO categories (id, name, icon, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)")
			_, err = h.DB.DB.Exec(insertQuery, catID, tCat.Name, "Globe", tCat.Description, time.Now(), time.Now())
			if err != nil {
				log.Printf("Error creating category %s: %v", tCat.Name, err)
				continue
			}
			countCategories++
		}

		for _, tPat := range tCat.Patterns {
			// 2. Check if pattern exists or create it
			var patID string
			query = h.DB.convertPlaceholders("SELECT id FROM patterns WHERE name = ? AND category_id = ?")
			err = h.DB.DB.QueryRow(query, tPat.Name, catID).Scan(&patID)

			if err == sql.ErrNoRows {
				patID = generateID()
				insertQuery := h.DB.convertPlaceholders("INSERT INTO patterns (id, category_id, name, icon, description, theory, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
				_, err = h.DB.DB.Exec(insertQuery, patID, catID, tPat.Name, "Code", tPat.Description, "", time.Now(), time.Now())
				if err != nil {
					log.Printf("Error creating pattern %s: %v", tPat.Name, err)
					continue
				}
				countPatterns++
			}

			for _, tProb := range tPat.MatchedProblems {
				// 3. Check if problem exists or create it
				var probID string
				// Use the ID from Thita if possible, otherwise generate one
				targetProbID := tProb.ID
				if targetProbID == "" {
					targetProbID = generateID()
				}

				query = h.DB.convertPlaceholders("SELECT id FROM problems WHERE title = ? AND pattern_id = ?")
				err = h.DB.DB.QueryRow(query, tProb.Title, patID).Scan(&probID)

				if err == sql.ErrNoRows {
					insertQuery := h.DB.convertPlaceholders("INSERT INTO problems (id, pattern_id, title, difficulty, description, input, output, constraints, sample_input, sample_output, explanation, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
					_, err = h.DB.DB.Exec(insertQuery,
						targetProbID, patID, tProb.Title, tProb.Difficulty,
						"Description pending fetch...", "See description", "See description",
						"No specific constraints provided.", "", "", "", "",
						time.Now(), time.Now())
					if err != nil {
						log.Printf("Error creating problem %s: %v", tProb.Title, err)
						continue
					}
					countProblems++
				}
			}
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Bulk fetch completed successfully",
		"stats": map[string]int{
			"categoriesCreated": countCategories,
			"patternsCreated":   countPatterns,
			"problemsCreated":   countProblems,
		},
	})
}

// ClearAllData wipes the entire database (except users)
func (h *Handlers) ClearAllData(w http.ResponseWriter, r *http.Request) {
	if isDemoUser(r) {
		respondWithError(w, http.StatusForbidden, "Demo users cannot clear data")
		return
	}

	// Order matters due to foreign keys
	tables := []string{"solutions", "problems", "patterns", "categories", "learning_resources", "roadmap_items", "learning_topics"}

	for _, table := range tables {
		_, err := h.DB.DB.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to clear table %s: %v", table, err))
			return
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "All data cleared successfully"})
}

// GenerateCategoryDescription uses AI to generate category description
func (h *Handlers) GenerateCategoryDescription(w http.ResponseWriter, r *http.Request) {
	if isDemoUser(r) {
		respondWithError(w, http.StatusForbidden, "Demo users cannot generate content")
		return
	}

	var req struct {
		Name   string `json:"name"`
		Prompt string `json:"prompt"` // Optional user prompt
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Category name is required")
		return
	}

	basePrompt := fmt.Sprintf("Generate a comprehensive description for a coding problem category named \"%s\". \n\n"+
		"The description should:\n"+
		"- Explain what types of problems belong to this category\n"+
		"- Describe the common characteristics and patterns\n"+
		"- Mention typical use cases\n"+
		"- Include a C++ code example demonstrating a typical problem in this category\n"+
		"- Be clear, concise, and informative\n"+
		"- Be written in markdown format with proper code blocks\n\n"+
		"Format the response in markdown. Include C++ code examples in code blocks marked as ```cpp. Return ONLY the markdown content, no JSON wrapper.", req.Name)

	// Add user's custom prompt if provided
	var prompt string
	if req.Prompt != "" && strings.TrimSpace(req.Prompt) != "" {
		prompt = fmt.Sprintf("%s\n\nAdditional instructions from user: %s", basePrompt, req.Prompt)
	} else {
		prompt = basePrompt
	}

	content, err := h.callAI(prompt)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating description: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"description": content})
}

// GeneratePatternContent uses AI to generate pattern description and theory
func (h *Handlers) GeneratePatternContent(w http.ResponseWriter, r *http.Request) {
	if isDemoUser(r) {
		respondWithError(w, http.StatusForbidden, "Demo users cannot generate content")
		return
	}

	var req struct {
		Name         string `json:"name"`
		CategoryName string `json:"categoryName"`
		ContentType  string `json:"contentType"` // "description" or "theory"
		Prompt       string `json:"prompt"`      // Optional user prompt
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Pattern name is required")
		return
	}

	var basePrompt string
	if req.ContentType == "theory" {
		basePrompt = fmt.Sprintf("Generate comprehensive theory content for the algorithm pattern \"%s\" in the category \"%s\".\n\n"+
			"The theory should include:\n"+
			"- Detailed explanation of the pattern\n"+
			"- When to use this pattern\n"+
			"- Step-by-step approach\n"+
			"- Time and space complexity analysis\n"+
			"- Common variations\n"+
			"- Example use cases with C++ code examples\n"+
			"- Visual explanations where helpful\n\n"+
			"IMPORTANT: You MUST include C++ code examples. All code examples must be in C++ language. Use code blocks marked as ```cpp for all C++ code.\n"+
			"Provide at least one complete, working C++ implementation example that demonstrates the pattern.\n"+
			"Format the response in markdown with proper headings. Return ONLY the markdown content, no JSON wrapper.", req.Name, req.CategoryName)
	} else {
		basePrompt = fmt.Sprintf("Generate a concise description for the algorithm pattern \"%s\" in the category \"%s\".\n\n"+
			"The description should:\n"+
			"- Briefly explain what this pattern is\n"+
			"- Mention key characteristics\n"+
			"- Give a quick overview of when to use it\n"+
			"- Include a simple C++ code snippet example\n"+
			"- Be clear and concise (2-3 sentences plus code example)\n\n"+
			"Format the response in markdown. Include a C++ code example in a code block marked as ```cpp. Return ONLY the markdown content, no JSON wrapper.", req.Name, req.CategoryName)
	}

	// Add user's custom prompt if provided
	var prompt string
	if req.Prompt != "" && strings.TrimSpace(req.Prompt) != "" {
		prompt = fmt.Sprintf("%s\n\nAdditional instructions from user: %s", basePrompt, req.Prompt)
	} else {
		prompt = basePrompt
	}

	content, err := h.callAI(prompt)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating content: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"content": content})
}

// callAI is a helper function to call the AI service
func (h *Handlers) callAI(prompt string) (string, error) {
	// Validate API key
	if h.AIAPIKey == "" || h.AIAPIKey == "your-api-key-here" {
		return "", fmt.Errorf("AI API key is not configured. Please set the AI_API_KEY environment variable or pass it via -ai-api-key flag")
	}

	openRouterURL := "https://openrouter.ai/api/v1/chat/completions"
	requestBody := map[string]interface{}{
		"model": "openai/gpt-4o-mini",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", openRouterURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+h.AIAPIKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/algovault")
	httpReq.Header.Set("X-Title", "AlgoVault")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to connect to AI service: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		// Try to parse error response for better error message
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Code    int    `json:"code"`
			} `json:"error"`
		}
		if json.Unmarshal(body, &errorResp) == nil && errorResp.Error.Message != "" {
			if resp.StatusCode == 401 {
				return "", fmt.Errorf("invalid or expired OpenRouter API key. Please check your AI_API_KEY. Error: %s", errorResp.Error.Message)
			}
			return "", fmt.Errorf("AI service error (%d): %s", resp.StatusCode, errorResp.Error.Message)
		}
		return "", fmt.Errorf("AI service error (status %d): %s", resp.StatusCode, string(body))
	}

	var openRouterResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		return "", err
	}

	if len(openRouterResp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI service")
	}

	content := openRouterResp.Choices[0].Message.Content
	// Remove outer markdown code blocks if the entire response is wrapped in them
	// But keep code blocks that are part of the content (like ```cpp)
	contentTrimmed := strings.TrimSpace(content)
	if strings.HasPrefix(contentTrimmed, "```") && strings.HasSuffix(contentTrimmed, "```") {
		// Check if it's a wrapper (starts with ```json or ```markdown or just ```)
		firstLine := strings.Split(contentTrimmed, "\n")[0]
		if strings.Contains(firstLine, "json") || strings.Contains(firstLine, "markdown") ||
			(strings.Count(contentTrimmed, "```") == 2 && strings.Index(contentTrimmed, "```") == 0) {
			// Remove the wrapper
			lines := strings.Split(contentTrimmed, "\n")
			if len(lines) > 2 {
				content = strings.Join(lines[1:len(lines)-1], "\n")
			}
		}
	}

	return content, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Learning Resource Handlers

func (h *Handlers) GetLearningTopics(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.DB.Query("SELECT id, name, icon, description, slug, created_at, updated_at FROM learning_topics ORDER BY name ASC")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}
	defer rows.Close()

	var topics []LearningTopic
	for rows.Next() {
		var t LearningTopic
		if err := rows.Scan(&t.ID, &t.Name, &t.Icon, &t.Description, &t.Slug, &t.CreatedAt, &t.UpdatedAt); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning topic")
			return
		}
		topics = append(topics, t)
	}

	respondWithJSON(w, http.StatusOK, topics)
}

func (h *Handlers) GetLearningTopicBySlug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	var t LearningTopic
	query := h.DB.convertPlaceholders("SELECT id, name, icon, description, slug, created_at, updated_at FROM learning_topics WHERE slug = ?")
	err := h.DB.DB.QueryRow(query, slug).Scan(&t.ID, &t.Name, &t.Icon, &t.Description, &t.Slug, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Topic not found")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, t)
}

func (h *Handlers) GetLearningResources(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicID := vars["topicId"]

	query := h.DB.convertPlaceholders("SELECT id, topic_id, title, content, type, url, order_index, created_at, updated_at FROM learning_resources WHERE topic_id = ? ORDER BY order_index ASC")
	rows, err := h.DB.DB.Query(query, topicID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}
	defer rows.Close()

	var resources []LearningResource
	for rows.Next() {
		var res LearningResource
		if err := rows.Scan(&res.ID, &res.TopicID, &res.Title, &res.Content, &res.Type, &res.URL, &res.OrderIndex, &res.CreatedAt, &res.UpdatedAt); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning resource")
			return
		}
		resources = append(resources, res)
	}

	respondWithJSON(w, http.StatusOK, resources)
}

func (h *Handlers) GetRoadmap(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicID := vars["topicId"]

	query := h.DB.convertPlaceholders("SELECT id, topic_id, title, description, order_index, status, created_at, updated_at FROM roadmap_items WHERE topic_id = ? ORDER BY order_index ASC")
	rows, err := h.DB.DB.Query(query, topicID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}
	defer rows.Close()

	var items []RoadmapItem
	for rows.Next() {
		var item RoadmapItem
		if err := rows.Scan(&item.ID, &item.TopicID, &item.Title, &item.Description, &item.OrderIndex, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning roadmap item")
			return
		}
		items = append(items, item)
	}

	respondWithJSON(w, http.StatusOK, items)
}
