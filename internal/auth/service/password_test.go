package service_test

import (
	"context"
	"testing"
	"time"

	"ptmuda.tech/backend/internal/auth/service"
)

func TestPasswordHashAndVerify(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	p := service.Passwords{}
	password := "SecretP@ssw0rd2026!"

	hash, err := p.Hash(ctx, password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	valid, err := p.Verify(ctx, hash, password)
	if err != nil {
		t.Fatalf("failed to verify password: %v", err)
	}
	if !valid {
		t.Fatalf("expected password verification to succeed")
	}

	invalid, err := p.Verify(ctx, hash, "WrongPassword!")
	if err != nil {
		t.Fatalf("unexpected error on wrong password: %v", err)
	}
	if invalid {
		t.Fatalf("expected wrong password verification to fail")
	}
}
