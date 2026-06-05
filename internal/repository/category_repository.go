package repository

import (
	"context"
	"database/sql"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetAll(ctx context.Context) ([]domain.Category, error) {
	const sqlStr = `
		SELECT id, parent_id, name
		FROM categories
		ORDER BY parent_id NULLS FIRST, name
	`
	rows, err := r.db.QueryContext(ctx, sqlStr)
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err, "failed to query categories")
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Name); err != nil {
			return nil, apperror.ErrInternal.Wrap(err, "failed to scan category")
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, apperror.ErrInternal.Wrap(err, "failed to iterate categories")
	}
	return categories, nil
}
