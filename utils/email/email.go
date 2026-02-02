package email

import (
	"fmt"
	"net/smtp"

	"vestra-ecommerce/config"
	"vestra-ecommerce/utils/logging"
)

var smtpCfg config.SMTPConfig // Holds SMTP configuration loaded from app config

// Init initializes SMTP configuration
func Init(cfg config.SMTPConfig) {
	smtpCfg = cfg
	logging.Debug.Printf("SMTP initialized with host: %s, user: %s\n", smtpCfg.Host, smtpCfg.Username)
}

// SendOTP sends an OTP email to the given recipient
func SendOTP(to string, otp string) error {

	from := smtpCfg.Username // SMTP authenticated sender address

	fromHeader := smtpCfg.From // Email From header value
	if fromHeader == "" {
		fromHeader = from
	}

	// RFC 5322 compliant email message
	msg := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: Your OTP for Vestra Ecommerce\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n"+
			"Hello!\n\nYour OTP is: %s\nIt will expire in 5 minutes.\n\nThanks,\nVestra Ecommerce Team",
		fromHeader, to, otp,
	)

	addr := fmt.Sprintf("%s:%d", smtpCfg.Host, smtpCfg.Port) // SMTP server address

	auth := smtp.PlainAuth("", smtpCfg.Username, smtpCfg.Password, smtpCfg.Host) // SMTP authentication

	// Send email using SMTP
	if err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg)); err != nil {
		logging.Error.Printf("Failed to send OTP to %s: %v\n", to, err)
		return err
	}

	logging.Debug.Printf("OTP sent successfully to %s\n", to)
	return nil
}
