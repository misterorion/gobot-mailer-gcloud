package GobotMailer

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/rs/zerolog/log"
)

var apiKey = os.Getenv("MG_API_KEY")
var domain = os.Getenv("MG_DOMAIN")
var authUser = os.Getenv("AUTH_USER")
var authPass = os.Getenv("AUTH_PASS")

type message struct {
	Name    string
	Email   string
	Comment string
	IP      string
}

func GobotMailer(w http.ResponseWriter, r *http.Request) {

	// Check if API key present
	if apiKey == "" || domain == "" {
		log.Error().Msg("Missing API key")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Setup CORS
	(w).Header().Set("Access-Control-Allow-Origin", "https://misterorion.com")
	(w).Header().Set("Access-Control-Allow-Methods", "OPTIONS, POST")
	(w).Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Content-Length")
	if r.Method == "OPTIONS" {
		return
	}

	// Try to get the user's IP address for logging purposes
	userIP := strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]

	// Authorize
	u, p, ok := r.BasicAuth()
	if !ok || u != authUser || p != authPass {
		log.Warn().
			Str("IP", userIP).
			Msg("Unauthorized attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Validate the HTTP request
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/contact/" {
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

	// Decode and Validate data received by the form
	var m message
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
	m.IP = userIP
	err = sendMail(m)
	if err != nil {
		log.Error().
			Str("Error", err.Error()).
			Msg("Email sending failed.")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Return a success response and log the message
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/text")
	_, err = w.Write([]byte("Success"))

	log.Info().
		Str("IP", m.IP).
		Str("Name", m.Name).
		Str("Email", m.Email).
		Str("Comment", m.Comment).
		Msg("Message sent")
}

// sendMail sends a message through the Mailgun API using an HTML template
func sendMail(m message) error {
	t, err := template.ParseFiles("serverless_function_source_code/template.html")
	if err != nil {
		return err
	}

	var data = map[string]string{
		"name":    m.Name,
		"email":   m.Email,
		"comment": m.Comment,
		"ip":      m.IP,
	}

	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, data); err != nil {
		return err
	}

	mg := mailgun.NewMailgun(domain, apiKey)

	message := mg.NewMessage(
		"Form GoBot <form-gobot@mechapower.com>", // From address
		"New form submission",                    // Subject
		"Enable html to read this message.",      // Plaintext body
		"orion@mechapower.com",                   // To
	)

	message.SetHtml(buffer.String())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, _, err = mg.Send(ctx, message)
	if err != nil {
		return err
	}

	return nil
}
