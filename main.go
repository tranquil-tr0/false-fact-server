// main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
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
// Standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// /health (health check) endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message":   "Server is running successfully!",
			"timestamp": time.Now(),
		},
	}
	json.NewEncoder(w).Encode(response)
}

// Handles root and unknown endpoints
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	response := APIResponse{
		Success: false,
		Error: map[string]interface{}{
			"message":   fmt.Sprintf("Endpoint '%s' does not exist", r.URL.Path),
			"timestamp": time.Now(),
		},
	}
	json.NewEncoder(w).Encode(response)
}

// /analyze/article endpoint handler
func analyzeArticleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req AnalyzeArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	result, err := AiAnalyzeArticle(req.Content, req.Title, req.URL, req.LastEdited, selectedModel)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   map[string]interface{}{"message": "AI analysis failed", "error": err.Error()},
		})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    result,
	})
}

// /analyze/text/long endpoint handler
func analyzeLongTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req AnalyzeTextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	result, err := AiAnalyzeTextLong(req.Content, selectedModel)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   map[string]interface{}{"message": "AI analysis failed", "error": err.Error()},
		})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    result,
	})
}

// /analyze/text/short endpoint handler
func analyzeShortTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req AnalyzeTextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	result, err := AiAnalyzeTextShort(req.Content, selectedModel)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   map[string]interface{}{"message": "AI analysis failed", "error": err.Error()},
		})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    result,
	})
}

var selectedModel Model

// CORS middleware
func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func main() {
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose debug output")
	flag.Parse()
	http.HandleFunc("/", withCORS(rootHandler))
	http.HandleFunc("/health", withCORS(healthHandler))
	http.HandleFunc("/analyze/article", withCORS(analyzeArticleHandler))
	http.HandleFunc("/analyze/text/short", withCORS(analyzeShortTextHandler))
	http.HandleFunc("/analyze/text/long", withCORS(analyzeLongTextHandler))

	err := godotenv.Load()
	if err != nil {
		log.Fatal("No .env file found or failed to load .env")
	}

	// Model selection via env file
	modelEnv := strings.ToLower(os.Getenv("MODEL"))
	switch modelEnv {
	case "pollinations":
		selectedModel = Pollinations
		fmt.Println("[main] Using model: Pollinations")
	case "gemini":
		selectedModel = Gemini
		fmt.Println("[main] Using model: Gemini")
		if os.Getenv("GEMINI_API_KEY") == "" {
			log.Fatal("GEMINI_API_KEY is not set in the environment. Please set it to use the Gemini model.")
		}
	case "":
		log.Fatal("No MODEL set in env file, please set MODEL to 'Pollinations' or 'Gemini'")
	default:
		log.Fatal(fmt.Sprintf("[main] Unknown MODEL '%s'", modelEnv))
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("No PORT variable set in env file\n")
	}
	port = ":" + port

	fmt.Printf("🚀 Server starting on port %s\n", port)
	fmt.Printf("📡 API endpoints:\n")
	fmt.Printf("   - POST /analyze/article\n")
	fmt.Printf("   - POST /analyze/text/short\n")
	fmt.Printf("   - POST /analyze/text/long\n")
	fmt.Printf("   - GET  /health\n")
	fmt.Printf("\n💡 Access your server at: http://localhost%s\n", port)
	if verbose {
		fmt.Printf("[main] Verbose mode enabled\n")
	}

	// Start the server
	log.Fatal(http.ListenAndServe(port, nil))
}
