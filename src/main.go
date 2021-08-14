package p

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
)

var apiKey string = os.Getenv("MG_API_KEY")
var domain string = os.Getenv("MG_DOMAIN")

var authUser string = os.Getenv("AUTH_USER")
var authPass string = os.Getenv("AUTH_PASS")

type Message struct {
	Name, Email, Comment, IP string
}

func Main(w http.ResponseWriter, r *http.Request) {

	userIp := r.Header.Get("X-Forwarded-For")

	u, p, ok := r.BasicAuth()

	if !ok || u != authUser || p != authPass {
		log.Warn().
			Str("IP", userIp).
			Msg("Unauthorized attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if apiKey == "" || domain == "" {
		log.Error().Msg("Missing API key")
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

	// Validate the JSON
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

	m.IP = userIp
	err = SendMail(m)
	// err = MockMail(m)

	if err != nil {
		log.Error().
			Str("Error", err.Error()).
			Msg("Email sending failed.")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Send OK status
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/text")
	w.Write([]byte("OK"))

	// Log the results
	log.Info().
		Str("IP", m.IP).
		Str("Name", m.Name).
		Str("Email", m.Email).
		Str("Comment", m.Comment).
		Msg("Message sent")
}
