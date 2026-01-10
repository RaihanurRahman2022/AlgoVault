package main

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	err := h.DB.DB.QueryRow(
		"SELECT id, email, name, password, role FROM users WHERE email = ?",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Password, &user.Role)

	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusUnauthorized, "User not found. Please register first or check your email.")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	// Compare password - check if password hash is valid first
	if len(user.Password) == 0 {
		respondWithError(w, http.StatusInternalServerError, "User account error")
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
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

// Register handles user registration - DISABLED
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusForbidden, "Registration is disabled. Please use the demo account: demo@algovault.com / demo123")
}

// Category handlers

func (h *Handlers) GetCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.DB.Query(`
		SELECT c.id, c.name, c.icon, c.description, c.created_at, c.updated_at,
		       COUNT(DISTINCT p.id) as pattern_count
		FROM categories c
		LEFT JOIN patterns p ON p.category_id = c.id
		GROUP BY c.id
		ORDER BY c.created_at DESC
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

	_, err := h.DB.DB.Exec(
		"INSERT INTO categories (id, name, icon, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		cat.ID, cat.Name, cat.Icon, cat.Description, cat.CreatedAt, cat.UpdatedAt,
	)
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
	_, err := h.DB.DB.Exec(
		"UPDATE categories SET name = ?, icon = ?, description = ?, updated_at = ? WHERE id = ?",
		cat.Name, cat.Icon, cat.Description, cat.UpdatedAt, id,
	)
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

	_, err := h.DB.DB.Exec("DELETE FROM categories WHERE id = ?", id)
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

	rows, err := h.DB.DB.Query(`
		SELECT p.id, p.category_id, p.name, p.icon, p.description, p.created_at, p.updated_at,
		       COUNT(DISTINCT pr.id) as problem_count
		FROM patterns p
		LEFT JOIN problems pr ON pr.pattern_id = p.id
		WHERE p.category_id = ?
		GROUP BY p.id
		ORDER BY p.created_at DESC
	`, categoryID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var patterns []Pattern
	for rows.Next() {
		var pat Pattern
		err := rows.Scan(&pat.ID, &pat.CategoryID, &pat.Name, &pat.Icon, &pat.Description, &pat.CreatedAt, &pat.UpdatedAt, &pat.ProblemCount)
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

	_, err := h.DB.DB.Exec(
		"INSERT INTO patterns (id, category_id, name, icon, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		pat.ID, pat.CategoryID, pat.Name, pat.Icon, pat.Description, pat.CreatedAt, pat.UpdatedAt,
	)
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
	_, err := h.DB.DB.Exec(
		"UPDATE patterns SET name = ?, icon = ?, description = ?, updated_at = ? WHERE id = ?",
		pat.Name, pat.Icon, pat.Description, pat.UpdatedAt, id,
	)
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

	_, err := h.DB.DB.Exec("DELETE FROM patterns WHERE id = ?", id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting pattern")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Pattern deleted"})
}

// Problem handlers

func (h *Handlers) GetProblems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	patternID := vars["patternId"]

	rows, err := h.DB.DB.Query(`
		SELECT id, pattern_id, title, difficulty, description, input, output, 
		       constraints, sample_input, sample_output, explanation, notes,
		       created_at, updated_at
		FROM problems
		WHERE pattern_id = ?
		ORDER BY created_at DESC
	`, patternID)
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
	err := h.DB.DB.QueryRow(`
		SELECT id, pattern_id, title, difficulty, description, input, output,
		       constraints, sample_input, sample_output, explanation, notes,
		       created_at, updated_at
		FROM problems
		WHERE id = ?
	`, id).Scan(
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

	_, err := h.DB.DB.Exec(`
		INSERT INTO problems (id, pattern_id, title, difficulty, description, input, output,
		                     constraints, sample_input, sample_output, explanation, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, prob.ID, prob.PatternID, prob.Title, prob.Difficulty,
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
		_, err := h.DB.DB.Exec(`
			INSERT INTO solutions (id, problem_id, language, code, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, sol.ID, sol.ProblemID, sol.Language, sol.Code, sol.CreatedAt, sol.UpdatedAt)
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
	_, err := h.DB.DB.Exec(`
		UPDATE problems SET title = ?, difficulty = ?, description = ?, input = ?, output = ?,
		                    constraints = ?, sample_input = ?, sample_output = ?, explanation = ?,
		                    notes = ?, updated_at = ?
		WHERE id = ?
	`, prob.Title, prob.Difficulty, prob.Description, prob.Input, prob.Output,
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
		err := h.DB.DB.QueryRow(`
			SELECT id FROM solutions WHERE problem_id = ? AND language = ?
		`, id, sol.Language).Scan(&existingID)
		
		if err == sql.ErrNoRows {
			// New solution - generate ID
			sol.ID = generateID()
			sol.CreatedAt = time.Now()
			_, err = h.DB.DB.Exec(`
				INSERT INTO solutions (id, problem_id, language, code, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?)
			`, sol.ID, sol.ProblemID, sol.Language, sol.Code, sol.CreatedAt, sol.UpdatedAt)
		} else if err == nil {
			// Existing solution - update it
			sol.ID = existingID
			_, err = h.DB.DB.Exec(`
				UPDATE solutions SET code = ?, updated_at = ? WHERE problem_id = ? AND language = ?
			`, sol.Code, sol.UpdatedAt, sol.ProblemID, sol.Language)
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

	_, err := h.DB.DB.Exec("DELETE FROM problems WHERE id = ?", id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting problem")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Problem deleted"})
}

// Helper function to get solutions for a problem
func (h *Handlers) getSolutions(problemID string) ([]Solution, error) {
	rows, err := h.DB.DB.Query(`
		SELECT id, problem_id, language, code, created_at, updated_at
		FROM solutions
		WHERE problem_id = ?
	`, problemID)
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
	Title         string `json:"title"`
	Difficulty    string `json:"difficulty"`
	Description   string `json:"description"`
	Input         string `json:"input"`
	Output        string `json:"output"`
	Constraints   string `json:"constraints"`
	SampleInput   string `json:"sampleInput"`
	SampleOutput  string `json:"sampleOutput"`
	Explanation   string `json:"explanation"`
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
