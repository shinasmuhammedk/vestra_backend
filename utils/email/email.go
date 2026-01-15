package email

import (
	"fmt"
	"log"
	"net/smtp"

	"vestra-ecommerce/config"
)

var smtpCfg config.SMTPConfig

// Init sets the SMTP config from app.yaml
func Init(cfg config.SMTPConfig) {
	smtpCfg = cfg
	log.Printf("[email] SMTP initialized with host: %s, user: %s\n", smtpCfg.Host, smtpCfg.Username)
}

// SendOTP sends an OTP email to the recipient
func SendOTP(to string, otp string) error {
	// Envelope sender (must match authenticated username)
	from := smtpCfg.Username

	// From header
	fromHeader := smtpCfg.From
	if fromHeader == "" {
		fromHeader = from
	}

	// RFC 5322-compliant message
	msg := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: Your OTP for Vestra Ecommerce\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n"+
			"Hello!\n\nYour OTP is: %s\nIt will expire in 5 minutes.\n\nThanks,\nVestra Ecommerce Team",
		fromHeader, to, otp,
	)

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", smtpCfg.Host, smtpCfg.Port)

	// Authentication
	auth := smtp.PlainAuth("", smtpCfg.Username, smtpCfg.Password, smtpCfg.Host)

	// Send email
	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
	if err != nil {
		log.Printf("[email] Failed to send OTP to %s: %v\n", to, err)
		return err
	}

	log.Printf("[email] OTP sent successfully to %s\n", to)
	return nil
}
