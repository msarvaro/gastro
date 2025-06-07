package repository

import (
	"context"
	"database/sql"
	"restaurant-management/internal/domain/entity"
)

type MenuRepository struct {
	db *sql.DB
}

func NewMenuRepository(db *sql.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

// Menu operations
func (r *MenuRepository) GetByID(ctx context.Context, id int) (*entity.Menu, error) {
	query := `SELECT id, business_id, name, description, is_active, created_at, updated_at 
	         FROM menus WHERE id = $1`

	menu := &entity.Menu{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&menu.ID, &menu.BusinessID, &menu.Name, &menu.Description,
		&menu.IsActive, &menu.CreatedAt, &menu.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Load categories for this menu
	categories, err := r.GetCategoriesByMenuID(ctx, menu.ID)
	if err != nil {
		return nil, err
	}
	menu.Categories = categories

	return menu, nil
}

func (r *MenuRepository) GetByBusinessID(ctx context.Context, businessID int) ([]*entity.Menu, error) {
	query := `SELECT id, business_id, name, description, is_active, created_at, updated_at 
	         FROM menus WHERE business_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []*entity.Menu
	for rows.Next() {
		menu := &entity.Menu{}
		err := rows.Scan(
			&menu.ID, &menu.BusinessID, &menu.Name, &menu.Description,
			&menu.IsActive, &menu.CreatedAt, &menu.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		menus = append(menus, menu)
	}

	return menus, nil
}

func (r *MenuRepository) GetActiveByBusinessID(ctx context.Context, businessID int) (*entity.Menu, error) {
	query := `SELECT id, business_id, name, description, is_active, created_at, updated_at 
	         FROM menus WHERE business_id = $1 AND is_active = true LIMIT 1`

	menu := &entity.Menu{}
	err := r.db.QueryRowContext(ctx, query, businessID).Scan(
		&menu.ID, &menu.BusinessID, &menu.Name, &menu.Description,
		&menu.IsActive, &menu.CreatedAt, &menu.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Load categories for this menu
	categories, err := r.GetCategoriesByMenuID(ctx, menu.ID)
	if err != nil {
		return nil, err
	}
	menu.Categories = categories

	return menu, nil
}

func (r *MenuRepository) Create(ctx context.Context, menu *entity.Menu) error {
	query := `INSERT INTO menus (business_id, name, description, is_active, created_at, updated_at)
	         VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		menu.BusinessID, menu.Name, menu.Description, menu.IsActive,
	).Scan(&menu.ID)

	return err
}

func (r *MenuRepository) Update(ctx context.Context, menu *entity.Menu) error {
	query := `UPDATE menus SET name = $1, description = $2, is_active = $3, updated_at = NOW()
	         WHERE id = $4`

	_, err := r.db.ExecContext(ctx, query,
		menu.Name, menu.Description, menu.IsActive, menu.ID,
	)

	return err
}

func (r *MenuRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM menus WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Category operations
func (r *MenuRepository) GetCategoriesByMenuID(ctx context.Context, menuID int) ([]*entity.Category, error) {
	query := `SELECT id, menu_id, business_id, name, created_at, updated_at 
	         FROM categories WHERE menu_id = $1 ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query, menuID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*entity.Category
	for rows.Next() {
		category := &entity.Category{}
		err := rows.Scan(
			&category.ID, &category.MenuID, &category.BusinessID,
			&category.Name, &category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Load dishes for this category
		dishes, err := r.GetDishesByCategoryID(ctx, category.ID)
		if err != nil {
			return nil, err
		}
		category.Dishes = dishes

		categories = append(categories, category)
	}

	return categories, nil
}

func (r *MenuRepository) CreateCategory(ctx context.Context, category *entity.Category) error {
	query := `INSERT INTO categories (menu_id, business_id, name, created_at, updated_at)
	         VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		category.MenuID, category.BusinessID, category.Name,
	).Scan(&category.ID)

	return err
}

func (r *MenuRepository) UpdateCategory(ctx context.Context, category *entity.Category) error {
	query := `UPDATE categories SET name = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, category.Name, category.ID)
	return err
}

func (r *MenuRepository) DeleteCategory(ctx context.Context, id int) error {
	query := `DELETE FROM categories WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Dish operations
func (r *MenuRepository) GetDishesByCategoryID(ctx context.Context, categoryID int) ([]*entity.Dish, error) {
	query := `SELECT id, category_id, business_id, name, description, price, image_url,
	         is_available, preparation_time, calories, allergens, created_at, updated_at
	         FROM dishes WHERE category_id = $1 ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dishes []*entity.Dish
	for rows.Next() {
		dish := &entity.Dish{}
		err := rows.Scan(
			&dish.ID, &dish.CategoryID, &dish.BusinessID, &dish.Name,
			&dish.Description, &dish.Price, &dish.ImageURL, &dish.IsAvailable,
			&dish.PreparationTime, &dish.Calories, &dish.Allergens,
			&dish.CreatedAt, &dish.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		dishes = append(dishes, dish)
	}

	return dishes, nil
}

func (r *MenuRepository) GetDishByID(ctx context.Context, id int) (*entity.Dish, error) {
	query := `SELECT id, category_id, business_id, name, description, price, image_url,
	         is_available, preparation_time, calories, allergens, created_at, updated_at
	         FROM dishes WHERE id = $1`

	dish := &entity.Dish{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&dish.ID, &dish.CategoryID, &dish.BusinessID, &dish.Name,
		&dish.Description, &dish.Price, &dish.ImageURL, &dish.IsAvailable,
		&dish.PreparationTime, &dish.Calories, &dish.Allergens,
		&dish.CreatedAt, &dish.UpdatedAt,
	)

	return dish, err
}

func (r *MenuRepository) CreateDish(ctx context.Context, dish *entity.Dish) error {
	query := `INSERT INTO dishes (category_id, business_id, name, description, price, 
	         image_url, is_available, preparation_time, calories, allergens, created_at, updated_at)
	         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()) RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		dish.CategoryID, dish.BusinessID, dish.Name, dish.Description, dish.Price,
		dish.ImageURL, dish.IsAvailable, dish.PreparationTime, dish.Calories,
		dish.Allergens,
	).Scan(&dish.ID)

	return err
}

func (r *MenuRepository) UpdateDish(ctx context.Context, dish *entity.Dish) error {
	query := `UPDATE dishes SET name = $1, description = $2, price = $3, image_url = $4,
	         is_available = $5, preparation_time = $6, calories = $7, allergens = $8, updated_at = NOW()
	         WHERE id = $9`

	_, err := r.db.ExecContext(ctx, query,
		dish.Name, dish.Description, dish.Price, dish.ImageURL,
		dish.IsAvailable, dish.PreparationTime, dish.Calories, dish.Allergens,
		dish.ID,
	)

	return err
}

func (r *MenuRepository) DeleteDish(ctx context.Context, id int) error {
	query := `DELETE FROM dishes WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *MenuRepository) SetDishAvailability(ctx context.Context, id int, isAvailable bool) error {
	query := `UPDATE dishes SET is_available = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, isAvailable, id)
	return err
}
