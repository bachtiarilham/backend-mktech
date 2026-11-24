package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"ptmuda.tech/backend/internal/apperror"
	"ptmuda.tech/backend/internal/auth/model"
	"ptmuda.tech/backend/internal/auth/repository"
	"ptmuda.tech/backend/internal/auth/service"
	coreModel "ptmuda.tech/backend/internal/core/model"
	"ptmuda.tech/backend/internal/infra/postgres"
)

type UseCase struct {
	repo      repository.Repository
	store     *postgres.Store
	passwords service.Passwords
	limiter   *service.Limiter
}

func New(repo repository.Repository, store *postgres.Store, passwords service.Passwords, limiter *service.Limiter) *UseCase {
	return &UseCase{
		repo:      repo,
		store:     store,
		passwords: passwords,
		limiter:   limiter,
	}
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func generateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (u *UseCase) Login(ctx context.Context, in model.LoginRequest, ip, userAgent string) (*model.LoginResponse, *model.TokenPair, error) {
	ident := strings.TrimSpace(in.Identifier)
	if ident == "" || in.Password == "" {
		return nil, nil, apperror.Invalid("Username/email dan password wajib diisi")
	}

	if err := u.limiter.Limit("login:"+ip+":"+strings.ToLower(ident), 6, time.Minute); err != nil {
		return nil, nil, err
	}

	record, err := u.repo.ByIdentifier(ctx, ident)
	if errors.Is(err, apperror.ErrNotFound) {
		_, _ = u.passwords.Verify(ctx, "", in.Password)
		return nil, nil, apperror.ErrUnauthorized
	}
	if err != nil {
		return nil, nil, err
	}

	passHash := coreModel.Text(record, "password_hash")
	valid, err := u.passwords.Verify(ctx, passHash, in.Password)
	if err != nil || !valid {
		return nil, nil, apperror.ErrUnauthorized
	}

	if !coreModel.Flag(record, "is_active") {
		return nil, nil, apperror.Public(apperror.ErrForbidden, "Akun admin nonaktif")
	}

	adminID := int64(coreModel.Number(record, "id"))
	deviceID := in.DeviceID
	if !coreModel.UUID(deviceID) {
		newDev, err := coreModel.NewID()
		if err != nil {
			return nil, nil, err
		}
		deviceID = newDev
	}

	accessToken, err := generateOpaqueToken()
	if err != nil {
		return nil, nil, err
	}
	refreshToken, err := generateOpaqueToken()
	if err != nil {
		return nil, nil, err
	}

	accessHash := hashToken(accessToken)
	refreshHash := hashToken(refreshToken)

	now := time.Now()
	accessExp := now.Add(15 * time.Minute)
	refreshExp := now.Add(7 * 24 * time.Hour)

	var sessionRecord coreModel.Record
	err = u.store.Do(ctx, func(txCtx context.Context) error {
		sr, err := u.repo.SaveSession(txCtx, adminID, deviceID, accessHash, refreshHash, userAgent, ip, accessExp, refreshExp)
		if err != nil {
			return err
		}
		sessionRecord = sr
		return u.repo.UpdateLastLogin(txCtx, adminID)
	})
	if err != nil {
		return nil, nil, err
	}

	_ = sessionRecord
	profile, err := u.Profile(ctx, adminID)
	if err != nil {
		return nil, nil, err
	}

	tokens := &model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AccessExpiry: accessExp,
	}

	resp := &model.LoginResponse{
		Admin:       *profile,
		AccessToken: accessToken,
		ExpiresIn:   900, // 15 minutes
	}

	return resp, tokens, nil
}

func (u *UseCase) Refresh(ctx context.Context, refreshToken, ip, userAgent string) (*model.TokenPair, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, apperror.ErrUnauthorized
	}

	refreshHash := hashToken(refreshToken)
	session, err := u.repo.FindByRefreshHash(ctx, refreshHash)
	if err != nil {
		return nil, apperror.ErrUnauthorized
	}

	sessionID := int64(coreModel.Number(session, "session_id"))
	adminID := int64(coreModel.Number(session, "admin_id"))
	deviceID := coreModel.Text(session, "device_id")

	newAccess, err := generateOpaqueToken()
	if err != nil {
		return nil, err
	}
	newRefresh, err := generateOpaqueToken()
	if err != nil {
		return nil, err
	}

	newAccessHash := hashToken(newAccess)
	newRefreshHash := hashToken(newRefresh)

	now := time.Now()
	accessExp := now.Add(15 * time.Minute)
	refreshExp := now.Add(7 * 24 * time.Hour)

	err = u.store.Do(ctx, func(txCtx context.Context) error {
		if err := u.repo.RevokeSession(txCtx, sessionID); err != nil {
			return err
		}
		_, err := u.repo.SaveSession(txCtx, adminID, deviceID, newAccessHash, newRefreshHash, userAgent, ip, accessExp, refreshExp)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &model.TokenPair{
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
		AccessExpiry: accessExp,
	}, nil
}

func (u *UseCase) Logout(ctx context.Context, sessionID int64) error {
	if sessionID <= 0 {
		return nil
	}
	return u.repo.RevokeSession(ctx, sessionID)
}

func (u *UseCase) Profile(ctx context.Context, adminID int64) (*model.AdminProfile, error) {
	r, err := u.repo.Profile(ctx, adminID)
	if err != nil {
		return nil, err
	}

	p := &model.AdminProfile{
		ID:        int64(coreModel.Number(r, "id")),
		Username:  coreModel.Text(r, "username"),
		Email:     coreModel.Text(r, "email"),
		FullName:  coreModel.Text(r, "full_name"),
		Role:      coreModel.Text(r, "role"),
		IsActive:  coreModel.Flag(r, "is_active"),
		CreatedAt: time.Now(),
	}

	if lt := coreModel.Text(r, "last_login_at"); lt != "" {
		if t, err := time.Parse(time.RFC3339, lt); err == nil {
			p.LastLoginAt = &t
		}
	}

	return p, nil
}

func (u *UseCase) AuthenticateAdmin(ctx context.Context, token string) (model.AdminIdentity, error) {
	if strings.TrimSpace(token) == "" {
		return model.AdminIdentity{}, apperror.ErrUnauthorized
	}

	h := hashToken(token)
	r, err := u.repo.FindByAccessHash(ctx, h)
	if err != nil {
		return model.AdminIdentity{}, apperror.ErrUnauthorized
	}

	return model.AdminIdentity{
		ID:        int64(coreModel.Number(r, "admin_id")),
		Username:  coreModel.Text(r, "username"),
		Email:     coreModel.Text(r, "email"),
		FullName:  coreModel.Text(r, "full_name"),
		Role:      coreModel.Text(r, "role"),
		IsActive:  coreModel.Flag(r, "is_active"),
		SessionID: int64(coreModel.Number(r, "session_id")),
	}, nil
}

func (u *UseCase) BootstrapAdmin(ctx context.Context, username, email, password, fullName, role string) (coreModel.Record, error) {
	hash, err := u.passwords.Hash(ctx, password)
	if err != nil {
		return nil, err
	}
	if role == "" {
		role = "SUPER_ADMIN"
	}
	return u.repo.BootstrapAdmin(ctx, username, email, hash, fullName, role)
}
