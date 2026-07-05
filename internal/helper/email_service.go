package helper

import (
	"bytes"
	"context"
	"html/template"
	"net/smtp"
	"shorter-url/internal/domain"
	"shorter-url/templates"
	"time"
)

type smtpEmailService struct {
	host string
	port string
	auth smtp.Auth
	from string
}

func NewEmailService(host, port, email, password string) domain.EmailService {
	auth := smtp.PlainAuth("", email, password, host)

	return &smtpEmailService{
		host: host,
		port: port,
		auth: auth,
		from: email,
	}
}

func (s *smtpEmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	message := []byte(
		"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-version: 1.0;\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\";\r\n" +
			"\r\n" +
			body + "\r\n")

	addrs := s.host + ":" + s.port
	return smtp.SendMail(addrs, s.auth, s.from, []string{to}, message)
}

func (s *smtpEmailService) SendEmailWithHTML(ctx context.Context, to string, context any, subjectText string, templateName string) error {
	from := "From: \"Shorter URL\" <" + s.from + ">\r\n"
	clientEmail := "To: " + to + "\r\n"
	date := "Date: " + time.Now().Format(time.RFC1123Z) + "\r\n"
	subject := "Subject: " + subjectText + "\r\n"
	mime := "MIME-version: 1.0;\r\n"
	content_type := "Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n"

	header := from + clientEmail + date + subject + mime + content_type

	template, err := template.ParseFS(templates.EmailTemplatesFS, "mail/"+templateName+".html")
	if err != nil {
		return err
	}

	var bodyBytes bytes.Buffer
	if err = template.Execute(&bodyBytes, context); err != nil {
		return err
	}
	bodyString := bodyBytes.String()

	mail := []byte(header + bodyString)
	address := s.host + ":" + s.port

	return smtp.SendMail(address, s.auth, s.from, []string{to}, mail)
}
