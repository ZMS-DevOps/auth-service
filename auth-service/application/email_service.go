package application

import (
	"bytes"
	"github.com/go-mail/mail"
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"html/template"
	"log"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

func (service *EmailService) GetVerificationCodeEmailBody(receiverEmail string, verification domain.Verification) string {
	t, _ := template.ParseFiles(domain.VerificationEmailTemplate)
	var body bytes.Buffer
	if err := service.executeEmail(receiverEmail, verification, t, &body); err != nil {
		return ""
	}

	return body.String()
}

func (service *EmailService) SendEmail(subject, body string) {
	m := mail.NewMessage()
	m.SetHeader("From", domain.SenderEmailAddress)
	m.SetHeader("To", domain.SenderEmailAddress)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := mail.NewDialer(domain.EmailHost, domain.EmailPort, domain.SenderEmailAddress, domain.AppPassword)

	if err := d.DialAndSend(m); err != nil {
		log.Fatal(err)
	}
}

func (service *EmailService) getVerificationCodeUrl(receiverEmail string, verification domain.Verification) string {
	return "http://localhost/app/booking/auth/verify/" + verification.Id.Hex() + "?email=" + receiverEmail + "&userId=" + verification.UserId
}

func (service *EmailService) executeEmail(receiverEmail string, verification domain.Verification, t *template.Template, body *bytes.Buffer) error {
	return t.Execute(body, struct {
		Code int
		Url  string
	}{
		Code: verification.Code,
		Url:  service.getVerificationCodeUrl(receiverEmail, verification),
	})
}
