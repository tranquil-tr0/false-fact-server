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

	result, err := AiAnalyzeArticle(req.Content, req.Title, req.URL, req.LastEdited, selectedModel)
	if err != nil {
		http.Error(w, `{"status":"error","data":"AI analysis failed", "error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// /analyze/text/long endpoint handler
func analyzeLongTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req AnalyzeTextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"status":"error","data":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	result, err := AiAnalyzeTextLong(req.Content, selectedModel)
	if err != nil {
		http.Error(w, `{"status":"error","data":"AI analysis failed", "error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// /analyze/text/short endpoint handler
func analyzeShortTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req AnalyzeTextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"status":"error","data":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	result, err := AiAnalyzeTextShort(req.Content, selectedModel)
	if err != nil {
		http.Error(w, `{"status":"error","data":"AI analysis failed", "error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

var selectedModel Model

func main() {
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose debug output")
	flag.Parse()
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/analyze/article", analyzeArticleHandler)
	http.HandleFunc("/analyze/text/short", analyzeShortTextHandler)
	http.HandleFunc("/analyze/text/long", analyzeLongTextHandler)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("No .env file found or failed to load .env")
	}

	// Model selection via env file
	modelEnv := strings.ToLower(os.Getenv("MODEL"))
	switch modelEnv {
	case "pollinations":
		selectedModel = Pollinations
		fmt.Println("[main] Using model: Pollinations (set by MODEL)")
	case "gemini":
		selectedModel = Gemini
		fmt.Println("[main] Using model: Gemini (default or set by MODEL)")
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
