package mailer_test

import (
	"testing"

	"github.com/leahgarrett/image-management-system/services/backend/internal/mailer"
)

func TestLogMailer_DoesNotError(t *testing.T) {
	m := mailer.NewLogMailer()
	err := m.SendMagicLink("test@example.com", "http://localhost:3000/auth/verify?token=abc123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}
