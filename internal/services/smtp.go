package services

// import (
// 	"fmt"
// 	"net/mail"
// 	"net/smtp"
// 	"strings"

// 	"github.com/Burmuley/ovoo/internal/config"
// )

// type SMTPClient struct {
// 	username string
// 	password string
// 	host     string
// 	port     int
// 	from     string
// 	auth     smtp.Auth
// }

// func NewSMTPClient(cfg config.MailNotificationConfig) (*SMTPClient, error) {
// 	if _, err := mail.ParseAddress(cfg.FromAddress); err != nil {
// 		return nil, err
// 	}

// 	fa := mail.Address{Name: cfg.FromName, Address: cfg.FromAddress}
// 	client := &SMTPClient{
// 		username: cfg.SMTPUsername,
// 		password: cfg.SMTPPassword,
// 		host:     cfg.SMTPHost,
// 		port:     cfg.SMTPPort,
// 		from:     fa.String(),
// 		auth:     nil,
// 	}
// 	// assume target SMTP gateway only supports plain auth schema (for now)
// 	if strings.TrimSpace(client.username) != "" && strings.TrimSpace(client.password) != "" {
// 		client.auth = smtp.PlainAuth("", client.username, client.password, client.host)
// 	}

// 	return client, nil
// }

// func (s *SMTPClient) SendMessage(to string, msg []byte) error {
// 	return smtp.SendMail(fmt.Sprintf("%s:%d", s.host, s.port), s.auth, s.from, []string{to}, msg)
// }
