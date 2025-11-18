package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/msteinert/pam"
	"log/slog"
	"time"
)

func AuthenticateWithPassword(username string, password []byte) (bool, error) {
	if len(password) == 0 {
		return false, errors.New("empty password")
	}

	pwCopy := make([]byte, len(password))
	copy(pwCopy, password)
	zero(password)

	conv := pam.ConversationFunc(func(style pam.Style, msg string) (string, error) {
		switch style {
		case pam.PromptEchoOff:
			return string(pwCopy), nil
		case pam.PromptEchoOn:
			return string(pwCopy), nil
		case pam.ErrorMsg, pam.TextInfo:
			return "", nil
		default:
			return "", nil
		}
	})

	tx, err := pam.Start("sshd", username, conv)
	if err != nil {
		slog.Error("failed to start PAM transaction", "user", username, "error", err)
		zero(pwCopy)
		return false, fmt.Errorf("pam start: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- tx.Authenticate(0)
	}()

	select {
	case err = <-done:
		if err != nil {
			slog.Error("PAM authenticate failed", "user", username, "error", err)
			return false, fmt.Errorf("authenticate: %w", err)
		}
		if err = tx.AcctMgmt(0); err != nil {
			slog.Info("PAM account management failed", "user", username, "error", err)
			return false, fmt.Errorf("acct-mgmt: %w", err)
		}

		return true, nil
	case <-ctx.Done():
		slog.Info("PAM authentication timed out", "user", username)
		return false, fmt.Errorf("authentication timed out")
	}
}

func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
