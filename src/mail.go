package p

import (
	"bytes"
	"context"
	"text/template"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

// MockMail returns a nil error for testing
func MockMail(m Message) error {
	return nil
}

// SendMail sends a message through the Mailgun API
func SendMail(m Message) error {
	err := ParseTemplateAndSend(m)
	return err
}

// ParseTemplateAndSend sends a mailgun message
func ParseTemplateAndSend(m Message) error {
	t, err := template.ParseFiles("serverless_function_source_code/template.html")
	// t, err := template.ParseFiles("../template.html")

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

	err = sendComplexMessage(domain, buffer.String(), apiKey)
	if err != nil {
		return err
	}
	return nil
}

func sendComplexMessage(domain, content string, apiKey string) error {
	mg := mailgun.NewMailgun(domain, apiKey)
	m := mg.NewMessage(
		"Form GoBot <form-gobot@mechapower.com>", // From address
		"New form submission",                    // Subject
		"Enable html to read this message.",      // Plaintext body
		"orion@mechapower.com",                   // To
	)
	m.SetHtml(content)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, _, err := mg.Send(ctx, m)
	return err
}
