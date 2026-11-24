package repository

import (
	"context"
	"time"

	"ptmuda.tech/backend/internal/core/model"
	"ptmuda.tech/backend/internal/infra/postgres"
)

type Repository interface {
	ByIdentifier(ctx context.Context, identifier string) (model.Record, error)
	LockAdmin(ctx context.Context, adminID int64) (model.Record, error)
	UpdateLastLogin(ctx context.Context, adminID int64) error
	SaveSession(ctx context.Context, adminID int64, deviceID string, accessHash, refreshHash, agent, ip string, accessExp, refreshExp time.Time) (model.Record, error)
	FindByAccessHash(ctx context.Context, accessHash string) (model.Record, error)
	FindByRefreshHash(ctx context.Context, refreshHash string) (model.Record, error)
	RevokeSession(ctx context.Context, sessionID int64) error
	RevokeAllSessions(ctx context.Context, adminID int64) error
	Profile(ctx context.Context, adminID int64) (model.Record, error)
	BootstrapAdmin(ctx context.Context, username, email, passwordHash, fullName, role string) (model.Record, error)
}

type authRepo struct {
	store *postgres.Store
}

func New(store *postgres.Store) Repository {
	return &authRepo{store: store}
}

func (r *authRepo) ByIdentifier(ctx context.Context, identifier string) (model.Record, error) {
	return r.store.One(ctx, "LoginAdminByIdentifier", identifier)
}

func (r *authRepo) LockAdmin(ctx context.Context, adminID int64) (model.Record, error) {
	return r.store.One(ctx, "LoginAdminLock", adminID)
}

func (r *authRepo) UpdateLastLogin(ctx context.Context, adminID int64) error {
	return r.store.Exec(ctx, "LoginAdminUpdateLastLogin", adminID)
}

func (r *authRepo) SaveSession(ctx context.Context, adminID int64, deviceID string, accessHash, refreshHash, agent, ip string, accessExp, refreshExp time.Time) (model.Record, error) {
	return r.store.One(ctx, "LoginAdminSaveSession", adminID, deviceID, accessHash, refreshHash, agent, ip, accessExp, refreshExp)
}

func (r *authRepo) FindByAccessHash(ctx context.Context, accessHash string) (model.Record, error) {
	return r.store.One(ctx, "LoginAdminFindSessionByAccess", accessHash)
}

func (r *authRepo) FindByRefreshHash(ctx context.Context, refreshHash string) (model.Record, error) {
	return r.store.One(ctx, "LoginAdminFindSessionByRefresh", refreshHash)
}

func (r *authRepo) RevokeSession(ctx context.Context, sessionID int64) error {
	return r.store.Exec(ctx, "LoginAdminRevokeSession", sessionID)
}

func (r *authRepo) RevokeAllSessions(ctx context.Context, adminID int64) error {
	return r.store.Exec(ctx, "LoginAdminRevokeAllForAdmin", adminID)
}

func (r *authRepo) Profile(ctx context.Context, adminID int64) (model.Record, error) {
	return r.store.One(ctx, "AdminProfile", adminID)
}

func (r *authRepo) BootstrapAdmin(ctx context.Context, username, email, passwordHash, fullName, role string) (model.Record, error) {
	return r.store.One(ctx, "BootstrapAdminCreate", username, email, passwordHash, fullName, role)
}
