package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// ... (код с метриками Prometheus остается без изменений) ...
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"path", "status_code"}, // Добавим status_code для большей информативности
	)
)

func init() {
	// Настраиваем Logrus для вывода в формате JSON
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	prometheus.MustRegister(httpRequestsTotal)
}

type GeminiRequest struct {
	Prompt string `json:"prompt"`
}

type GeminiResponse struct {
	Response string `json:"response"`
}

func callGeminiAPI(prompt string) (string, error) {
	return "Это симулированный ответ от Gemini для: " + prompt, nil
}

func geminiHandler(w http.ResponseWriter, r *http.Request) {
	// Логируем начало обработки запроса
	log.WithFields(log.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
		"ip":     r.RemoteAddr,
	}).Info("Получен новый запрос")

	var req GeminiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Error("Ошибка декодирования JSON")
		http.Error(w, err.Error(), http.StatusBadRequest)
		httpRequestsTotal.WithLabelValues(r.URL.Path, "400").Inc()
		return
	}

	resp, err := callGeminiAPI(req.Prompt)
	if err != nil {
		log.WithError(err).Error("Ошибка при вызове Gemini API")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		httpRequestsTotal.WithLabelValues(r.URL.Path, "500").Inc()
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GeminiResponse{Response: resp})
	httpRequestsTotal.WithLabelValues(r.URL.Path, "200").Inc()
	log.WithFields(log.Fields{"prompt": req.Prompt}).Info("Запрос успешно обработан")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/gemini", geminiHandler).Methods("POST")
	r.Handle("/metrics", promhttp.Handler())

	log.Info("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
