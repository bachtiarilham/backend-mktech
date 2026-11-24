// Package service provides password hashing and rate limiting.
package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"ptmuda.tech/backend/internal/apperror"
)

type Passwords struct{}

var passwordJobs = make(chan struct{}, 2)

func acquirePassword(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case passwordJobs <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (Passwords) Hash(ctx context.Context, password string) (string, error) {
	if len(password) < 8 || len(password) > 256 {
		return "", apperror.Invalid("Password must contain 8 to 256 characters")
	}
	if err := acquirePassword(ctx); err != nil {
		return "", err
	}
	defer func() { <-passwordJobs }()

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	key := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)
	b := base64.RawStdEncoding
	return fmt.Sprintf("argon2id$v=19$m=65536,t=3,p=2$%s$%s", b.EncodeToString(salt), b.EncodeToString(key)), nil
}

func (Passwords) Verify(ctx context.Context, hash, password string) (bool, error) {
	if err := acquirePassword(ctx); err != nil {
		return false, err
	}
	defer func() { <-passwordJobs }()

	parts := strings.Split(hash, "$")
	if len(parts) != 5 || parts[0] != "argon2id" || parts[1] != "v=19" || parts[2] != "m=65536,t=3,p=2" || len(password) > 256 {
		argon2.IDKey([]byte("dummy"), []byte("dummy-salt-12345"), 3, 65536, 2, 32)
		return false, nil
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, nil
	}
	expectedKey, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, nil
	}

	key := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)
	return subtle.ConstantTimeCompare(key, expectedKey) == 1, nil
}
