package email

import (
	"context"
	"fmt"
)

type MockEmailService struct{}

// SendEmail implements services.EmailService.
func (*MockEmailService) SendEmail(
	ctx context.Context,
	to string,
	subject string,
	body string,
) {
	fmt.Printf("Sending email to %s with subject %s\n\n%s\n", to, subject, body)
}

// var _ services.EmailService = (*MockEmailService)(nil)
