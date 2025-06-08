package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"restaurant-management/internal/domain/menu"
	"strings"

	"github.com/lib/pq"
)

type MenuRepository struct {
	db *DB
}

func NewMenuRepository(db *DB) menu.Repository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) GetMenuItems(ctx context.Context, categoryID *int, businessID int) ([]menu.MenuItem, error) {
	// Check if description column exists
	var hasDescriptionColumn bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'dishes' AND column_name = 'description'
		)`).Scan(&hasDescriptionColumn)

	if err != nil {
		return nil, fmt.Errorf("checking for description column: %w", err)
	}

	// Build the query dynamically based on column existence
	query := `
		SELECT id, name, price, category_id, image_url, is_available, 
		       COALESCE(preparation_time, 0), COALESCE(calories, 0), allergens, `

	if hasDescriptionColumn {
		query += `COALESCE(description, ''), `
	}

	query += `COALESCE(business_id, 0), created_at, updated_at
		FROM dishes
		WHERE ($1::int IS NULL OR category_id = $1)
		ORDER BY category_id, name`

	rows, err := r.db.QueryContext(ctx, query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("querying menu items: %w", err)
	}
	defer rows.Close()

	var items []menu.MenuItem
	for rows.Next() {
		var item menu.MenuItem

		// Prepare scan destinations
		scanDest := []interface{}{
			&item.ID,
			&item.Name,
			&item.Price,
			&item.CategoryID,
			&item.ImageURL,
			&item.IsAvailable,
			&item.PreparationTime,
			&item.Calories,
			pq.Array(&item.Allergens),
		}

		// Add description to scan destinations only if column exists
		if hasDescriptionColumn {
			scanDest = append(scanDest, &item.Description)
		}

		// Add remaining scan destinations
		scanDest = append(scanDest,
			&item.BusinessID,
			&item.CreatedAt,
			&item.UpdatedAt)

		if err := rows.Scan(scanDest...); err != nil {
			return nil, fmt.Errorf("scanning menu item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return items, nil
}

func (r *MenuRepository) GetMenuItemByID(ctx context.Context, id int, businessID int) (*menu.MenuItem, error) {

	// Check if description column exists
	var hasDescriptionColumn bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'dishes' AND column_name = 'description'
		)`).Scan(&hasDescriptionColumn)

	if err != nil {
		return nil, fmt.Errorf("checking for description column: %w", err)
	}

	// Build the query dynamically based on column existence
	query := `
		SELECT id, name, price, category_id, image_url, is_available, 
		       COALESCE(preparation_time, 0), COALESCE(calories, 0), allergens, `

	if hasDescriptionColumn {
		query += `COALESCE(description, ''), `
	}

	query += `COALESCE(business_id, 0), created_at, updated_at
		FROM dishes
		WHERE id = $1 AND business_id = $2`

	// Prepare scan destinations
	var item menu.MenuItem
	scanDest := []interface{}{
		&item.ID,
		&item.Name,
		&item.Price,
		&item.CategoryID,
		&item.ImageURL,
		&item.IsAvailable,
		&item.PreparationTime,
		&item.Calories,
		pq.Array(&item.Allergens),
	}

	// Add description to scan destinations only if column exists
	if hasDescriptionColumn {
		scanDest = append(scanDest, &item.Description)
	}

	// Add remaining scan destinations
	scanDest = append(scanDest,
		&item.BusinessID,
		&item.CreatedAt,
		&item.UpdatedAt)

	err = r.db.QueryRowContext(ctx, query, id, businessID).Scan(scanDest...)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // Item not found
	}
	if err != nil {
		return nil, fmt.Errorf("scanning menu item by ID: %w", err)
	}
	return &item, nil
}

func (r *MenuRepository) CreateMenuItem(ctx context.Context, item menu.MenuItemCreate) (*menu.MenuItem, error) {
	// Check if description column exists
	var hasDescriptionColumn bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'dishes' AND column_name = 'description'
		)`).Scan(&hasDescriptionColumn)

	if err != nil {
		return nil, fmt.Errorf("checking for description column: %w", err)
	}

	// Build the query dynamically based on column existence
	query := `
		INSERT INTO dishes (name, price, category_id, image_url, is_available, preparation_time, calories, allergens, `

	if hasDescriptionColumn {
		query += `description, `
	}

	query += `business_id, created_at, updated_at)
		VALUES (`

	// Create placeholders for parameters
	params := []interface{}{
		item.Name,
		item.Price,
		item.CategoryID,
		item.ImageURL,
		item.IsAvailable,
		nilOrVal(item.PreparationTime),
		nilOrVal(item.Calories),
		pq.Array(item.Allergens),
	}

	// Start with $1
	paramIndex := 1
	placeholders := []string{
		fmt.Sprintf("$%d", paramIndex),   // name
		fmt.Sprintf("$%d", paramIndex+1), // price
		fmt.Sprintf("$%d", paramIndex+2), // category_id
		fmt.Sprintf("$%d", paramIndex+3), // image_url
		fmt.Sprintf("$%d", paramIndex+4), // is_available
		fmt.Sprintf("$%d", paramIndex+5), // preparation_time
		fmt.Sprintf("$%d", paramIndex+6), // calories
		fmt.Sprintf("$%d", paramIndex+7), // allergens
	}
	paramIndex += 8

	// Add description placeholder only if column exists
	if hasDescriptionColumn {
		params = append(params, item.Description)
		placeholders = append(placeholders, fmt.Sprintf("$%d", paramIndex))
		paramIndex++
	}

	// Add remaining parameters
	params = append(params, nilOrVal(item.BusinessID))
	placeholders = append(placeholders, fmt.Sprintf("$%d", paramIndex)) // business_id
	paramIndex++

	// Add placeholders for NOW() timestamps
	placeholders = append(placeholders, "NOW()", "NOW()")

	// Complete the query
	query += strings.Join(placeholders, ", ") + `)
		RETURNING id, name, price, category_id, image_url, is_available, 
		       COALESCE(preparation_time, 0), COALESCE(calories, 0), allergens, `

	if hasDescriptionColumn {
		query += `COALESCE(description, ''), `
	}

	query += `COALESCE(business_id, 0), created_at, updated_at`

	// Prepare scan destinations
	var created menu.MenuItem
	scanDest := []interface{}{
		&created.ID,
		&created.Name,
		&created.Price,
		&created.CategoryID,
		&created.ImageURL,
		&created.IsAvailable,
		&created.PreparationTime,
		&created.Calories,
		pq.Array(&created.Allergens),
	}

	// Add description to scan destinations only if column exists
	if hasDescriptionColumn {
		scanDest = append(scanDest, &created.Description)
	}

	// Add remaining scan destinations
	scanDest = append(scanDest,
		&created.BusinessID,
		&created.CreatedAt,
		&created.UpdatedAt)

	// Execute the query
	err = r.db.QueryRowContext(ctx, query, params...).Scan(scanDest...)

	if err != nil {
		return nil, fmt.Errorf("creating menu item: %w", err)
	}
	return &created, nil
}

func (r *MenuRepository) UpdateMenuItem(ctx context.Context, id int, item menu.MenuItemUpdate) (*menu.MenuItem, error) {
	// Check if description column exists
	var hasDescriptionColumn bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'dishes' AND column_name = 'description'
		)`).Scan(&hasDescriptionColumn)

	if err != nil {
		return nil, fmt.Errorf("checking for description column: %w", err)
	}

	// Build a dynamic SET clause based on the fields that are provided
	setClauses := []string{}
	params := []interface{}{}
	paramCounter := 1

	// Add clauses only for fields that need to be updated
	if item.Name != "" {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", paramCounter))
		params = append(params, item.Name)
		paramCounter++
	}

	if item.Price > 0 {
		setClauses = append(setClauses, fmt.Sprintf("price = $%d", paramCounter))
		params = append(params, item.Price)
		paramCounter++
	}

	if item.CategoryID > 0 {
		setClauses = append(setClauses, fmt.Sprintf("category_id = $%d", paramCounter))
		params = append(params, item.CategoryID)
		paramCounter++
	}

	if item.ImageURL != "" {
		setClauses = append(setClauses, fmt.Sprintf("image_url = $%d", paramCounter))
		params = append(params, item.ImageURL)
		paramCounter++
	}

	if item.IsAvailable != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_available = $%d", paramCounter))
		params = append(params, *item.IsAvailable)
		paramCounter++
	}

	if item.PreparationTime > 0 {
		setClauses = append(setClauses, fmt.Sprintf("preparation_time = $%d", paramCounter))
		params = append(params, item.PreparationTime)
		paramCounter++
	}

	if item.Calories > 0 {
		setClauses = append(setClauses, fmt.Sprintf("calories = $%d", paramCounter))
		params = append(params, item.Calories)
		paramCounter++
	}

	if len(item.Allergens) > 0 {
		setClauses = append(setClauses, fmt.Sprintf("allergens = $%d", paramCounter))
		params = append(params, pq.Array(item.Allergens))
		paramCounter++
	}

	// Add description only if the column exists
	if hasDescriptionColumn && item.Description != "" {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", paramCounter))
		params = append(params, item.Description)
		paramCounter++
	}

	if item.BusinessID > 0 {
		setClauses = append(setClauses, fmt.Sprintf("business_id = $%d", paramCounter))
		params = append(params, item.BusinessID)
		paramCounter++
	}

	// Always update the timestamp
	setClauses = append(setClauses, "updated_at = NOW()")

	// If nothing to update, return the item as is
	if len(setClauses) == 1 { // Only the timestamp
		// Get the item and return it
		return r.GetMenuItemByID(ctx, id, item.BusinessID)
	}

	// Complete the query with the SET clauses
	query := fmt.Sprintf(`
		UPDATE dishes
		SET %s
		WHERE id = $%d
		RETURNING id, name, price, category_id, image_url, is_available, 
		       COALESCE(preparation_time, 0), COALESCE(calories, 0), allergens`,
		strings.Join(setClauses, ", "),
		paramCounter)

	// Add id to params
	params = append(params, id)

	// Add description to RETURNING clause only if it exists
	if hasDescriptionColumn {
		query += ", COALESCE(description, '')"
	}

	query += `, COALESCE(business_id, 0), created_at, updated_at`

	// Prepare scan destinations
	var updated menu.MenuItem
	scanDest := []interface{}{
		&updated.ID,
		&updated.Name,
		&updated.Price,
		&updated.CategoryID,
		&updated.ImageURL,
		&updated.IsAvailable,
		&updated.PreparationTime,
		&updated.Calories,
		pq.Array(&updated.Allergens),
	}

	// Add description to scan destinations only if column exists
	if hasDescriptionColumn {
		scanDest = append(scanDest, &updated.Description)
	}

	// Add remaining scan destinations
	scanDest = append(scanDest,
		&updated.BusinessID,
		&updated.CreatedAt,
		&updated.UpdatedAt)

	// Execute the query
	err = r.db.QueryRowContext(ctx, query, params...).Scan(scanDest...)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("updating menu item: %w - SQL: %s, Params: %v", err, query, params)
	}
	return &updated, nil
}

// Helper function to handle nil integer values
func nilOrVal(val int) interface{} {
	if val == 0 {
		return nil
	}
	return val
}

func (r *MenuRepository) DeleteMenuItem(ctx context.Context, id int, businessID int) error {
	query := `DELETE FROM dishes WHERE id = $1 AND business_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, businessID)
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

func (r *MenuRepository) GetCategories(ctx context.Context, businessID int) ([]menu.Category, error) {
	query := `
		SELECT id, name, business_id, created_at, updated_at
		FROM categories
		WHERE business_id = $1
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []menu.Category
	for rows.Next() {
		var category menu.Category
		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.BusinessID,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (r *MenuRepository) GetCategoryByID(ctx context.Context, id int, businessID int) (*menu.Category, error) {
	query := `
		SELECT id, name, business_id, created_at, updated_at
		FROM categories
		WHERE id = $1 AND business_id = $2`

	var category menu.Category
	err := r.db.QueryRowContext(ctx, query, id, businessID).Scan(
		&category.ID,
		&category.Name,
		&category.BusinessID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *MenuRepository) CreateCategory(ctx context.Context, category menu.CategoryCreate) (*menu.Category, error) {
	query := `
		INSERT INTO categories (name, business_id, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, name, business_id, created_at, updated_at`

	var created menu.Category
	err := r.db.QueryRowContext(ctx, query,
		category.Name,
		category.BusinessID,
	).Scan(
		&created.ID,
		&created.Name,
		&created.BusinessID,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		log.Println("Error creating category:", err)
		log.Println("Query:", query)
		log.Println("Params:", category.Name, category.BusinessID)
		return nil, err
	}
	return &created, nil
}

func (r *MenuRepository) UpdateCategory(ctx context.Context, id int, category menu.CategoryUpdate) (*menu.Category, error) {
	query := `
		UPDATE categories
		SET name = COALESCE($1, name),
			business_id = COALESCE($2, business_id),
			updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, business_id, created_at, updated_at`

	var updated menu.Category
	err := r.db.QueryRowContext(ctx, query,
		category.Name,
		nilOrVal(category.BusinessID),
		id,
	).Scan(
		&updated.ID,
		&updated.Name,
		&updated.BusinessID,
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

func (r *MenuRepository) DeleteCategory(ctx context.Context, id int, businessID int) error {
	query := `DELETE FROM categories WHERE id = $1 AND business_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, businessID)
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

// GetDishByID retrieves a specific dish by its ID
func (r *MenuRepository) GetDishByID(ctx context.Context, id int) (*menu.MenuItem, error) {
	// Check if description column exists
	var hasDescriptionColumn bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'dishes' AND column_name = 'description'
		)`).Scan(&hasDescriptionColumn)
	if err != nil {
		return nil, fmt.Errorf("checking for description column: %w", err)
	}

	var item menu.MenuItem

	query := `
		SELECT id, name, price, category_id, image_url, is_available, 
		       COALESCE(preparation_time, 0), COALESCE(calories, 0), allergens`

	// Add description to query only if column exists
	if hasDescriptionColumn {
		query += `, COALESCE(description, '')`
	}

	query += `, COALESCE(business_id, 0), created_at, updated_at
		FROM dishes 
		WHERE id = $1`

	// Prepare scan destinations
	scanDest := []interface{}{
		&item.ID,
		&item.Name,
		&item.Price,
		&item.CategoryID,
		&item.ImageURL,
		&item.IsAvailable,
		&item.PreparationTime,
		&item.Calories,
		pq.Array(&item.Allergens),
	}

	// Add description to scan destinations only if column exists
	if hasDescriptionColumn {
		scanDest = append(scanDest, &item.Description)
	}

	// Add remaining scan destinations
	scanDest = append(scanDest,
		&item.BusinessID,
		&item.CreatedAt,
		&item.UpdatedAt)

	err = r.db.QueryRowContext(ctx, query, id).Scan(scanDest...)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("dish with ID %d not found", id)
		}
		return nil, fmt.Errorf("getting dish by ID: %w", err)
	}

	return &item, nil
}
