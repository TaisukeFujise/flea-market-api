package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Register(ctx context.Context, user domain.User) error {
	const sqlStr = `
		INSERT INTO users (
			id,
			display_name,
			avatar_url
		) VALUES ($1, $2, $3)
	`
	result, err := r.db.ExecContext(ctx, sqlStr, user.ID, user.DisplayName, user.AvatarURL)
	if err != nil {
		if pq.As(err, pqerror.UniqueViolation) != nil {
			return apperror.ErrConflict.New("user already registered")
		}
		return apperror.ErrInternal.Wrap(err, "failed to insert user")
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return apperror.ErrInternal.New("insert user: unexpected rows affected")
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, id string, userUpdate domain.UserUpdate) error {
	const sqlStr = `
		UPDATE users SET
			display_name = COALESCE($2, display_name),
			avatar_url = COALESCE($3, avatar_url),
			updated_at = NOW()
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, sqlStr, id, userUpdate.DisplayName, userUpdate.AvatarURL)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to exec update user")
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return apperror.ErrInternal.New("update user: unexpected rows affected")
	}
	return nil
}

func (r *UserRepository) Get(ctx context.Context, id string) (domain.User, error) {
	const sqlStr = `
		SELECT id, display_name, avatar_url, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var user domain.User
	err := r.db.QueryRowContext(ctx, sqlStr, id).Scan(
		&user.ID,
		&user.DisplayName,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, apperror.ErrNotFound.Wrap(err, "user not found")
		}
		return domain.User{}, apperror.ErrInternal.Wrap(err, "failed to get user")
	}
	return user, nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	const sqlStr = `
		UPDATE users SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, sqlStr, id)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to delete user")
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return apperror.ErrNotFound.New("user not found")
	}
	return nil
}
