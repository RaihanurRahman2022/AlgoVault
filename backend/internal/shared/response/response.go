package response

import (
	"encoding/json"
	"net/http"
)

// Error sends a JSON error response
func Error(w http.ResponseWriter, code int, message string) {
	setCORSHeaders(w)
	JSON(w, code, map[string]string{"error": message})
}

// JSON sends a JSON response
func JSON(w http.ResponseWriter, code int, payload interface{}) {
	setCORSHeaders(w)

	response, err := json.Marshal(payload)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Error encoding response")
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
