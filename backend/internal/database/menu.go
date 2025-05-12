package database

import (
	"context"
	"database/sql"
	"restaurant-management/internal/models"
)

type MenuRepository struct {
	db *sql.DB
}

func NewMenuRepository(db *sql.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) GetMenuItems(ctx context.Context, categoryID *int) ([]models.MenuItem, error) {
	query := `
		SELECT id, name, price, category_id, image_url, is_available, created_at, updated_at
		FROM dishes
		WHERE ($1::int IS NULL OR category_id = $1)
		ORDER BY category_id, name`

	rows, err := r.db.QueryContext(ctx, query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.MenuItem
	for rows.Next() {
		var item models.MenuItem
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Price,
			&item.CategoryID,
			&item.ImageURL,
			&item.IsAvailable,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *MenuRepository) GetMenuItemByID(ctx context.Context, id int) (*models.MenuItem, error) {
	query := `
		SELECT id, name, price, category_id, image_url, is_available, created_at, updated_at
		FROM dishes
		WHERE id = $1`

	var item models.MenuItem
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.Name,
		&item.Price,
		&item.CategoryID,
		&item.ImageURL,
		&item.IsAvailable,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *MenuRepository) CreateMenuItem(ctx context.Context, item models.MenuItemCreate) (*models.MenuItem, error) {
	query := `
		INSERT INTO dishes (name, price, category_id, image_url, is_available, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, name, price, category_id, image_url, is_available, created_at, updated_at`

	var created models.MenuItem
	err := r.db.QueryRowContext(ctx, query,
		item.Name,
		item.Price,
		item.CategoryID,
		item.ImageURL,
		item.IsAvailable,
	).Scan(
		&created.ID,
		&created.Name,
		&created.Price,
		&created.CategoryID,
		&created.ImageURL,
		&created.IsAvailable,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *MenuRepository) UpdateMenuItem(ctx context.Context, id int, item models.MenuItemUpdate) (*models.MenuItem, error) {
	query := `
		UPDATE dishes
		SET name = COALESCE($1, name),
			price = COALESCE($2, price),
			category_id = COALESCE($3, category_id),
			image_url = COALESCE($4, image_url),
			is_available = COALESCE($5, is_available),
			updated_at = NOW()
		WHERE id = $6
		RETURNING id, name, price, category_id, image_url, is_available, created_at, updated_at`

	var updated models.MenuItem
	err := r.db.QueryRowContext(ctx, query,
		item.Name,
		item.Price,
		item.CategoryID,
		item.ImageURL,
		item.IsAvailable,
		id,
	).Scan(
		&updated.ID,
		&updated.Name,
		&updated.Price,
		&updated.CategoryID,
		&updated.ImageURL,
		&updated.IsAvailable,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *MenuRepository) DeleteMenuItem(ctx context.Context, id int) error {
	query := `DELETE FROM dishes WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *MenuRepository) GetCategories(ctx context.Context) ([]models.Category, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM categories
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (r *MenuRepository) CreateCategory(ctx context.Context, category models.CategoryCreate) (*models.Category, error) {
	query := `
		INSERT INTO categories (name, created_at, updated_at)
		VALUES ($1, NOW(), NOW())
		RETURNING id, name, created_at, updated_at`

	var created models.Category
	err := r.db.QueryRowContext(ctx, query,
		category.Name,
	).Scan(
		&created.ID,
		&created.Name,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *MenuRepository) UpdateCategory(ctx context.Context, id int, category models.CategoryUpdate) (*models.Category, error) {
	query := `
		UPDATE categories
		SET name = COALESCE($1, name),
			updated_at = NOW()
		WHERE id = $2
		RETURNING id, name, created_at, updated_at`

	var updated models.Category
	err := r.db.QueryRowContext(ctx, query,
		category.Name,
		id,
	).Scan(
		&updated.ID,
		&updated.Name,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *MenuRepository) DeleteCategory(ctx context.Context, id int) error {
	query := `DELETE FROM categories WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
