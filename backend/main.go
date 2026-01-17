package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// Configuration
	// Use PORT environment variable if available (for Render, Railway, etc.)
	portEnv := getEnv("PORT", "")
	if portEnv == "" {
		portEnv = "8080"
	}

	dbPath := flag.String("db", "./algovault.db", "Path to SQLite database file")
	port := flag.String("port", portEnv, "Server port")
	jwtSecret := flag.String("jwt-secret", getEnv("JWT_SECRET", "your-secret-key-change-in-production"), "JWT secret key")
	aiAPIKey := flag.String("ai-api-key", getEnv("AI_API_KEY", "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"), "OpenRouter API key")
	flag.Parse()

	// Log configuration
	log.Printf("Starting server...")
	log.Printf("PORT environment variable: %s", os.Getenv("PORT"))
	log.Printf("Using port: %s", *port)
	log.Printf("Database path: %s", *dbPath)
	if os.Getenv("DATABASE_URL") != "" {
		log.Printf("Using PostgreSQL (DATABASE_URL is set)")
	} else {
		log.Printf("Using SQLite")
	}

	// Initialize handlers (DB will be set once ready)
	handlers := &Handlers{
		DB:        nil,
		JWTSecret: *jwtSecret,
		AIAPIKey:  *aiAPIKey,
	}

	// Setup router
	router := mux.NewRouter()

	// CORS middleware - must be first
	router.Use(corsMiddleware)

	// Initialize database in background
	var db *Database
	var dbMutex sync.RWMutex
	dbReady := false
	dbReadyChan := make(chan bool, 1)

	go func() {
		var err error
		db, err = NewDatabase(*dbPath)
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		dbMutex.Lock()
		handlers.DB = db
		dbReady = true
		dbMutex.Unlock()
		dbReadyChan <- true
		log.Printf("Database ready")
	}()

	checkDBReady := func() bool {
		dbMutex.RLock()
		defer dbMutex.RUnlock()
		return dbReady && handlers.DB != nil
	}

	// Public routes with DB check
	router.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if !checkDBReady() {
			respondWithError(w, http.StatusServiceUnavailable, "Database initializing, please wait...")
			return
		}
		handlers.Login(w, r)
	}).Methods("POST", "OPTIONS")

	router.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		if !checkDBReady() {
			respondWithError(w, http.StatusServiceUnavailable, "Database initializing, please wait...")
			return
		}
		handlers.Register(w, r)
	}).Methods("POST", "OPTIONS")

	// Protected routes with DB check
	api := router.PathPrefix("/api").Subrouter()
	api.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !checkDBReady() {
				respondWithError(w, http.StatusServiceUnavailable, "Database initializing, please wait...")
				return
			}
			AuthMiddleware(*jwtSecret, db)(next).ServeHTTP(w, r)
		})
	})

	// Category routes
	api.HandleFunc("/categories", handlers.GetCategories).Methods("GET", "OPTIONS")
	api.HandleFunc("/categories", handlers.CreateCategory).Methods("POST", "OPTIONS")
	api.HandleFunc("/categories/{id}", handlers.UpdateCategory).Methods("PUT", "OPTIONS")
	api.HandleFunc("/categories/{id}", handlers.DeleteCategory).Methods("DELETE", "OPTIONS")

	// Pattern routes
	api.HandleFunc("/categories/{categoryId}/patterns", handlers.GetPatterns).Methods("GET", "OPTIONS")
	api.HandleFunc("/categories/{categoryId}/patterns", handlers.CreatePattern).Methods("POST", "OPTIONS")
	api.HandleFunc("/patterns/{id}", handlers.UpdatePattern).Methods("PUT", "OPTIONS")
	api.HandleFunc("/patterns/{id}", handlers.DeletePattern).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/patterns/{id}/theory", handlers.UpdatePatternTheory).Methods("PUT", "OPTIONS")

	// Problem routes
	api.HandleFunc("/patterns/{patternId}/problems", handlers.GetProblems).Methods("GET", "OPTIONS")
	api.HandleFunc("/patterns/{patternId}/problems", handlers.CreateProblem).Methods("POST", "OPTIONS")
	api.HandleFunc("/problems/{id}", handlers.GetProblem).Methods("GET", "OPTIONS")
	api.HandleFunc("/problems/{id}", handlers.UpdateProblem).Methods("PUT", "OPTIONS")
	api.HandleFunc("/problems/{id}", handlers.DeleteProblem).Methods("DELETE", "OPTIONS")

	// AI routes
	api.HandleFunc("/ai/generate-problem", handlers.GenerateProblem).Methods("POST", "OPTIONS")
	api.HandleFunc("/ai/generate-category-description", handlers.GenerateCategoryDescription).Methods("POST", "OPTIONS")
	api.HandleFunc("/ai/generate-pattern-content", handlers.GeneratePatternContent).Methods("POST", "OPTIONS")

	// Health check - returns OK immediately so Render can detect the port
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Always return OK for health check so Render can detect the port
		// The actual API will check DB readiness
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Debug endpoint to check if user exists (remove in production)
	router.HandleFunc("/api/debug/user-exists", func(w http.ResponseWriter, r *http.Request) {
		if !checkDBReady() {
			respondWithError(w, http.StatusServiceUnavailable, "Database initializing")
			return
		}
		email := r.URL.Query().Get("email")
		if email == "" {
			respondWithError(w, http.StatusBadRequest, "Email parameter required")
			return
		}

		var userID string
		var storedEmail string
		err := db.DB.QueryRow("SELECT id, email FROM users WHERE email = ? OR LOWER(email) = LOWER(?)", email, email).Scan(&userID, &storedEmail)
		if err == sql.ErrNoRows {
			respondWithJSON(w, http.StatusOK, map[string]interface{}{
				"exists":  false,
				"message": "User not found",
			})
			return
		}
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Database error: "+err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"exists": true,
			"id":     userID,
			"email":  storedEmail,
		})
	}).Methods("GET", "OPTIONS")

	// Start server immediately (so Render can detect port)
	addr := fmt.Sprintf("0.0.0.0:%s", *port)
	log.Printf("Server starting on %s", addr)

	// Cleanup on shutdown
	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	// Start server in goroutine so we can wait for DB
	serverErr := make(chan error, 1)
	go func() {
		if err := http.ListenAndServe(addr, router); err != nil {
			serverErr <- err
		}
	}()

	// Wait a moment for server to start
	time.Sleep(100 * time.Millisecond)
	log.Printf("Server listening, waiting for database...")

	// Wait for database to be ready
	<-dbReadyChan
	log.Printf("Database ready, server fully operational")

	// Wait for server error
	if err := <-serverErr; err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// corsMiddleware handles CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getEnv gets environment variable or returns default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
