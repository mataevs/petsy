package mailer

import (
	"fmt"

	"appengine"
	"appengine/mail"
)

type Message struct {
	To      []string
	Sender  string
	Subject string
	Body    string
}

func NewMessage(to []string, sender string, subject string, body string) *Message {
	return &Message{
		To:      to,
		Sender:  sender,
		Subject: subject,
		Body:    body,
	}
}

func SendEmail(c appengine.Context, to []string, sender string, subject string, body string) error {
	msg := &mail.Message{
		Sender:  sender,
		To:      to,
		Subject: subject,
		Body:    body,
	}

	if err := mail.Send(c, msg); err != nil {
		return fmt.Errorf("Could not send email: %v", err)
	}
	return nil
}
