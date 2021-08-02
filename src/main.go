package p

import (
	"encoding/json"
	"net/http"
	"os"

	"go.uber.org/zap"
)

var apiKey string = os.Getenv("MG_API_KEY")
var domain string = os.Getenv("MG_DOMAIN")

type Message struct {
	Name, Email, Comment, IP string
}

func Main(w http.ResponseWriter, r *http.Request) {

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if apiKey == "" || domain == "" {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Setup CORS headers
	(w).Header().Set("Access-Control-Allow-Origin", "https://misterorion.com")
	(w).Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	(w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length")

	// Validate the HTTP request
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.ContentLength > 800 {
		http.Error(w, "Content-Length too large", http.StatusBadRequest)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type not application/json", http.StatusBadRequest)
		return
	}

	// Validate the json
	var m Message
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, "Malformed JSON", http.StatusBadRequest)
		return
	}
	if len(m.Name) > 100 || len(m.Email) > 100 || len(m.Comment) > 500 {
		http.Error(w, "Length exceeded", http.StatusBadRequest)
		return
	}
	if m.Name == "" || m.Email == "" || m.Comment == "" {
		http.Error(w, "Missing JSON field", http.StatusBadRequest)
		return
	}

	// Send the message
	userIp := r.Header.Get("X-Forwarded-For")
	m.IP = userIp
	// err = SendMail(m)
	err = MockMail(m)

	if err != nil {
		logger.Error("Mail sending failed",
			zap.String("error", err.Error()),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Send OK status
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/text")
	w.Write([]byte("OK"))

	// Log the results
	logger.Info("Message sent",
		zap.String("name", m.Name),
		zap.String("comment", m.Comment),
		zap.String("email", m.Email),
		zap.String("ip", m.IP),
	)
}
