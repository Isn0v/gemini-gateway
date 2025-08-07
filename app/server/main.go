package main

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genai"
)

var (
	// geminiClient is the client for interacting with the Gemini API.
	geminiClient *genai.Client

	// geminiCtx is the context used to make requests to the Gemini API.
	geminiCtx = context.Background()

	// runningPlatform is the platform on which the server is running (local or cloud).
	runningPlatform = os.Getenv("RUNNING_PLATFORM")

	// httpRequestsTotal tracks the total number of HTTP requests received.
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"path", "status_code"}, // Added status_code for more informative metrics
	)
)

func init() {
	// Configure Logrus to output logs in JSON format
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil && runningPlatform == "local" {
		log.Fatal("Error loading .env file")
	}

	log.Info("Initializing Gemini client...")
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" && runningPlatform != "docker" {
		log.Fatal("GEMINI_API_KEY environment variable is not set")
	} else if runningPlatform == "docker" {
		log.Info("Running in Docker, skipping .env file loading and entering api key manually")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		apiKey = input[:len(input)-1] // Remove the newline character
		if err != nil {
			panic(err)
		}
	}

	geminiClient, err = genai.NewClient(geminiCtx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize Gemini client")
	} else {
		log.Info("Gemini client initialized successfully")
	}

	prometheus.MustRegister(httpRequestsTotal)
}

// GeminiRequest represents the structure of a request to the Gemini API.
type GeminiRequest struct {
	Prompt string `json:"prompt"`
}

// GeminiResponse represents the structure of a response from the Gemini API.
type GeminiResponse struct {
	Response string `json:"response"`
}

// callGeminiAPI sends a request to the Gemini API and retrieves the generated content.
func callGeminiAPI(ctx context.Context, prompt string) (string, error) {
	log.WithField("prompt", prompt).Info("Sending request to Gemini API")

	resp, err := geminiClient.Models.GenerateContent(
		ctx,
		"gemini-2.5-pro",
		genai.Text(prompt),
		nil,
	)

	if err != nil {
		log.WithError(err).Error("Error from Gemini API")
		return "", err
	}

	responseText := resp.Text()

	if responseText == "" {
		log.Warn("Gemini API returned an empty response")
		return "The model did not generate a response.", nil
	} else {
		log.WithField("response", responseText).Info("Received response from Gemini API")
	}

	return responseText, nil
}

// geminiHandler handles incoming HTTP requests to the /gemini endpoint.
func geminiHandler(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
		"ip":     r.RemoteAddr,
	}).Info("Received a new request")

	var req GeminiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Error("Error decoding JSON")
		http.Error(w, err.Error(), http.StatusBadRequest)
		httpRequestsTotal.WithLabelValues(r.URL.Path, "400").Inc()
		return
	}

	resp, err := callGeminiAPI(r.Context(), req.Prompt)
	if err != nil {
		log.WithError(err).Error("Error calling Gemini API")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		httpRequestsTotal.WithLabelValues(r.URL.Path, "500").Inc()
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GeminiResponse{Response: resp})
	httpRequestsTotal.WithLabelValues(r.URL.Path, "200").Inc()
	log.WithFields(log.Fields{"prompt": req.Prompt}).Info("Request processed successfully")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/gemini", geminiHandler).Methods("POST")
	r.Handle("/metrics", promhttp.Handler())

	log.Info("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
