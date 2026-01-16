package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "userID"
const roleKey contextKey = "role"

// AuthMiddleware validates JWT tokens and adds user ID to context
func AuthMiddleware(jwtSecret string, db *Database) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Ensure CORS headers are set before any processing
			setCORSHeaders(w)
			
			// Skip auth for OPTIONS requests (CORS preflight)
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondWithError(w, http.StatusUnauthorized, "Authorization header required")
				return
			}

			// Extract token from "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}

			tokenString := parts[1]
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				respondWithError(w, http.StatusUnauthorized, "Invalid token claims")
				return
			}

			userID, ok := claims["userID"].(string)
			if !ok {
				respondWithError(w, http.StatusUnauthorized, "Invalid user ID in token")
				return
			}

			// Get user role from claims (if available) or fetch from database
			userRole, _ := claims["role"].(string)
			if userRole == "" {
				// Fetch role from database
				var role string
				err := db.DB.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
				if err == nil {
					userRole = role
				} else {
					userRole = "admin" // Default to admin if not found
				}
			}

			// Add user ID and role to context
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, contextKey("role"), userRole)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// getUserID extracts user ID from context
func getUserID(r *http.Request) string {
	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}

// getUserRole extracts user role from context
func getUserRole(r *http.Request) string {
	role, ok := r.Context().Value(roleKey).(string)
	if !ok {
		return "admin" // Default to admin
	}
	return role
}

// isDemoUser checks if the current user is a demo user
func isDemoUser(r *http.Request) bool {
	return getUserRole(r) == "demo"
}

// respondWithError sends a JSON error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	// Ensure CORS headers are set on error responses too
	setCORSHeaders(w)
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	// Ensure CORS headers are set
	setCORSHeaders(w)
	
	response, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error encoding response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// setCORSHeaders sets CORS headers on the response
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}
