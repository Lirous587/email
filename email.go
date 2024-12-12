package email

import (
	"fmt"
	"net/smtp"
)

type Email struct {
	Host     string
	Port     string
	From     string
	Password string
	To       []string
	Subject  string
	Text     string
	Html     string
	body     []byte
	pool     *ConnectionPool
}

func NewEmail(pool *ConnectionPool) *Email {
	return &Email{pool: pool}
}

func (e *Email) Send(auth smtp.Auth) error {
	c, err := e.pool.Get()
	if err != nil {
		return fmt.Errorf("failed to get connection from pool: %w", err)
	}
	defer e.pool.Put(c)

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				return fmt.Errorf("failed to authenticate: %w", err)
			}
		}
	}

	if err = c.Mail(e.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, addr := range e.To {
		if err = c.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	body, err := e.write()
	if err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	_, err = w.Write(body)
	if err != nil {
		return fmt.Errorf("failed to write body: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}

func (e *Email) write() ([]byte, error) {
	headers := make(map[string]string)
	headers["From"] = e.From
	headers["Subject"] = e.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "multipart/alternative; boundary=boundary"

	var message string
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	message += "\r\n--boundary\r\n"
	message += "Content-Type: text/plain; charset=UTF-8\r\n\r\n"
	message += e.Text + "\r\n"

	message += "--boundary\r\n"
	message += "Content-Type: text/html; charset=UTF-8\r\n\r\n"
	message += e.Html + "\r\n"

	message += "--boundary--"

	return []byte(message), nil
}
