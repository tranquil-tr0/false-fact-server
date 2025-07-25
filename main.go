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

// /api endpoint
func apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		response := Response{
			Data:      "GET request received successfully",
			Timestamp: time.Now(),
			Status:    StatusSuccess,
		}
		json.NewEncoder(w).Encode(response)
	case "POST":
		response := Response{
			Data:      "POST request received successfully",
			Timestamp: time.Now(),
			Status:    StatusSuccess,
		}
		json.NewEncoder(w).Encode(response)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		response := Response{
			Data:      "Method not allowed",
			Timestamp: time.Now(),
			Status:    StatusError,
		}
		json.NewEncoder(w).Encode(response)
	}
}

// Handles root and unknown endpoints
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// non-existent endpoint
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		response := Response{
			Data:      fmt.Sprintf("Endpoint '%s' does not exist", r.URL.Path),
			Timestamp: time.Now(),
			Status:    StatusError,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// root endpoint
	response := Response{
		Data:      "Welcome to the False Fact Server API",
		Timestamp: time.Now(),
		Status:    StatusOnline,
	}
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/api", apiHandler)
	http.HandleFunc("/health", healthHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3088"
	}
	port = ":" + port

	fmt.Printf("🚀 Server starting on port %s\n", port)
	fmt.Printf("📡 API endpoints:\n")
	fmt.Printf("   - GET  / (root)\n")
	fmt.Printf("   - GET  /api\n")
	fmt.Printf("   - POST /api\n")
	fmt.Printf("   - GET  /health\n")
	fmt.Printf("\n💡 Access your server at: http://localhost%s\n", port)
	
	// Start the server
	log.Fatal(http.ListenAndServe(port, nil))
}
