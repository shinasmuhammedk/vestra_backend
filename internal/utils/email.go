package utils

import (
	"fmt"
	"net/smtp"
)

// SendOTPEmail sends the OTP to the specified email
func SendOTPEmail(toEmail, otp string) error {
	// Sender email credentials
	from := "shinaskdk@gmail.com"         // replace with your Gmail
	password := "ymbk isia oigo nsbc"        // Gmail App Password

	// SMTP server configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Email content
	subject := "Your Vestra OTP Code"
	body := fmt.Sprintf("Your OTP code is: %s. It expires in 5 minutes.", otp)

	message := []byte("Subject: " + subject + "\r\n\r\n" + body)

	// Authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Send email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, message)
	if err != nil {
		return err
	}

	fmt.Println("OTP sent to:", toEmail)
	return nil
}
