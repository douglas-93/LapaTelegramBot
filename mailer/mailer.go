package mailer

import (
	"fmt"
	"os"

	"github.com/wneessen/go-mail"
)

// Client representa um cliente de email
type Client struct {
	smtpServer   string
	smtpUser     string
	smtpPassword string
}

// NewClient cria um novo cliente de email a partir das variáveis de ambiente
func NewClient() *Client {
	return &Client{
		smtpServer:   os.Getenv("SMTP_SERVER"),
		smtpUser:     os.Getenv("SMTP_USER"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
	}
}

// EmailMessage representa uma mensagem de email
type EmailMessage struct {
	From        string
	To          []string
	Subject     string
	Body        string
	HTMLBody    string
	Attachments []string
}

// SendEmail envia um email
func (c *Client) SendEmail(msg EmailMessage) error {
	if c.smtpServer == "" || c.smtpUser == "" || c.smtpPassword == "" {
		return fmt.Errorf("configurações SMTP não definidas")
	}

	message := mail.NewMsg()

	// Define remetente
	if err := message.From(msg.From); err != nil {
		return fmt.Errorf("erro ao definir remetente: %w", err)
	}

	// Define destinatários
	if err := message.To(msg.To...); err != nil {
		return fmt.Errorf("erro ao definir destinatários: %w", err)
	}

	// Define assunto
	message.Subject(msg.Subject)

	// Define corpo do email
	if msg.HTMLBody != "" {
		message.SetBodyString(mail.TypeTextHTML, msg.HTMLBody)
	} else if msg.Body != "" {
		message.SetBodyString(mail.TypeTextPlain, msg.Body)
	}

	// Adiciona anexos
	for _, attachment := range msg.Attachments {
		message.AttachFile(attachment)
	}

	// Cria cliente SMTP
	client, err := mail.NewClient(
		c.smtpServer,
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(c.smtpUser),
		mail.WithPassword(c.smtpPassword),
	)
	if err != nil {
		return fmt.Errorf("erro ao criar cliente SMTP: %w", err)
	}

	// Envia email
	if err := client.DialAndSend(message); err != nil {
		return fmt.Errorf("erro ao enviar email: %w", err)
	}

	return nil
}
