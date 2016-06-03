package vain

import (
	"bytes"
	"fmt"
	"net/mail"
	"net/smtp"
)

// A Mailer is a type that knows how to send smtp mail.
type Mailer interface {
	Send(to mail.Address, subject, msg string) error
}

// NewEmail returns *Email struct to be able to send smtp
// or an error if it can't correctly parse the email address.
func NewEmail(from, host string, port int) (*Email, error) {
	if _, err := mail.ParseAddress(from); err != nil {
		return nil, fmt.Errorf("can't parse an email address for 'from': %v", err)
	}
	r := &Email{
		host: host,
		port: port,
		from: from,
	}
	return r, nil
}

// Email stores information required to use smtp.
type Email struct {
	host string
	port int
	from string
}

// Send sends a smtp email using the host and port in the Email struct and
//returns an error if there was a problem sending the email.
func (e Email) Send(to mail.Address, subject, msg string) error {
	c, err := smtp.Dial(fmt.Sprintf("%s:%d", e.host, e.port))
	if err != nil {
		return fmt.Errorf("couldn't dial mail server: %v", err)
	}
	defer c.Close()
	if err := c.Mail(e.from); err != nil {
		return err
	}
	if err := c.Rcpt(to.String()); err != nil {
		return err
	}
	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("problem sending mail: %v", err)
	}
	buf := bytes.NewBufferString("Subject: " + subject + "\n\n" + msg)
	buf.WriteTo(wc)
	if err := c.Quit(); err != nil {
		return nil
	}
	return err
}

type mockMail struct {
	msg string
}

func (m *mockMail) Send(to mail.Address, subject, msg string) error {
	m.msg = msg
	return nil
}
