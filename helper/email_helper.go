package helper

import (
	"fmt"

	"ticketing/config"

	gomail "gopkg.in/gomail.v2"
)

func SendEmail(to string, subject string, body string) error {
	if config.C.Env == "dev" {
		fmt.Println("[DEV MODE] Email to", to)
		fmt.Println("Subject:", subject)
		fmt.Println("Body:", body)
		return nil
	}
	// if prod mode is selected, send emails, continue to codes below

	m := gomail.NewMessage()
	m.SetHeader("From", config.C.FromEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(config.C.SMTPHost, config.C.SMTPPort, config.C.SMTPUser, config.C.SMTPPass)
	return d.DialAndSend(m)
}

func SendEmailOTP(to, username, otp string) error {
	subject := "QuickTix - One-Time Password (OTP)"
	body := fmt.Sprintf(
		"Hi %s,\n\n"+
			"Here is your One-Time Password (OTP) for completing your verification:\n\n"+
			"OTP Code: %s\n\n"+
			"This code will expire in 10 minutes.\n"+
			"Please do not share this code with anyone for your security.\n\n"+
			"If you did not request this, please ignore this email.\n\n"+
			"Thank you,\n"+
			"QuickTix",
		username, otp,
	)
	return SendEmail(to, subject, body)
}

func SendPaymentEmail(to, body string) error {
	subject := "QuickTix - Payment"
	return SendEmail(to, subject, body)
}
