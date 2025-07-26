// main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// possible status values
type StatusType string

const (
	StatusHealthy StatusType = "healthy"
	StatusSuccess StatusType = "success"
	StatusError   StatusType = "error"
	StatusOnline  StatusType = "online"
)

// API response structure
type Response struct {
	Data      string     `json:"data"`
	Timestamp time.Time  `json:"timestamp"`
	Status    StatusType `json:"status"`
}

// /health (health check) endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := Response{
		Data:      "Server is running successfully!",
		Timestamp: time.Now(),
		Status:    StatusHealthy,
	}
	json.NewEncoder(w).Encode(response)
}

// Handles root and unknown endpoints
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusNotFound)
	response := Response{
		Data:      fmt.Sprintf("Endpoint '%s' does not exist", r.URL.Path),
		Timestamp: time.Now(),
		Status:    StatusError,
	}
	json.NewEncoder(w).Encode(response)
	return
}

// /analyze/article endpoint handler
func analyzeArticleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req AnalyzeArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"status":"error","data":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	result, err := CallAIAnalyzeArticle(req.Content, req.Title, req.URL, req.LastEdited, Gemini)
	if err != nil {
		http.Error(w, `{"status":"error","data":"AI analysis failed"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/analyze/article", analyzeArticleHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3088"
	}
	port = ":" + port

	fmt.Printf("🚀 Server starting on port %s\n", port)
	fmt.Printf("📡 API endpoints:\n")
	fmt.Printf("   - GET  /api\n")
	fmt.Printf("   - POST /api\n")
	fmt.Printf("   - GET  /health\n")
	fmt.Printf("\n💡 Access your server at: http://localhost%s\n", port)

	// Start the server
	log.Fatal(http.ListenAndServe(port, nil))
}
