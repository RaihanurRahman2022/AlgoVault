package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

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

	// Initialize database
	db, err := NewDatabase(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize handlers
	handlers := &Handlers{
		DB:        db,
		JWTSecret: *jwtSecret,
		AIAPIKey:  *aiAPIKey,
	}

	// Setup router
	router := mux.NewRouter()

	// CORS middleware - must be first
	router.Use(corsMiddleware)

	// Public routes
	router.HandleFunc("/api/login", handlers.Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/register", handlers.Register).Methods("POST", "OPTIONS")

	// Protected routes
	api := router.PathPrefix("/api").Subrouter()
	api.Use(AuthMiddleware(*jwtSecret, db))

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

	// Problem routes
	api.HandleFunc("/patterns/{patternId}/problems", handlers.GetProblems).Methods("GET", "OPTIONS")
	api.HandleFunc("/patterns/{patternId}/problems", handlers.CreateProblem).Methods("POST", "OPTIONS")
	api.HandleFunc("/problems/{id}", handlers.GetProblem).Methods("GET", "OPTIONS")
	api.HandleFunc("/problems/{id}", handlers.UpdateProblem).Methods("PUT", "OPTIONS")
	api.HandleFunc("/problems/{id}", handlers.DeleteProblem).Methods("DELETE", "OPTIONS")
	
	// AI routes
	api.HandleFunc("/ai/generate-problem", handlers.GenerateProblem).Methods("POST", "OPTIONS")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Start server
	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Server starting on port %s", *port)
	log.Printf("Database: %s", *dbPath)
	log.Fatal(http.ListenAndServe(addr, router))
}

// corsMiddleware handles CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
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
