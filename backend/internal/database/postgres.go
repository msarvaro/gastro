package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"restaurant-management/configs"
	"restaurant-management/internal/middleware"
	"restaurant-management/internal/models"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	*sql.DB
}

func NewDB(config *configs.Config) (*DB, error) {
	db, err := sql.Open("postgres", config.GetDBConnString())
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}

	// Используем временные переменные для сканирования NULL значений
	var createdAt, updatedAt sql.NullTime
	var lastActive sql.NullTime

	err := db.QueryRow(`
        SELECT id, username, email, password, role, status, 
               COALESCE(created_at, CURRENT_TIMESTAMP) as created_at,
               COALESCE(updated_at, CURRENT_TIMESTAMP) as updated_at,
               COALESCE(last_active, CURRENT_TIMESTAMP) as last_active
        FROM users 
        WHERE username = $1`,
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Status,
		&createdAt,
		&updatedAt,
		&lastActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("No user found with username: %s", username)
			return nil, fmt.Errorf("user not found")
		}
		log.Printf("Database error: %v", err)
		return nil, err
	}

	// Присваиваем значения из NullTime
	user.CreatedAt = createdAt.Time
	user.UpdatedAt = updatedAt.Time
	user.LastActive = lastActive.Time

	return user, nil
}

func (db *DB) GetUserByID(id int, businessID int) (*models.User, error) {
	user := &models.User{}
	var name sql.NullString

	// First get the requesting user's role
	var requestingUserRole string
	err := db.QueryRow(`
		SELECT role FROM users WHERE id = $1`, id).Scan(&requestingUserRole)
	if err != nil {
		return nil, err
	}

	// Modify query based on user role
	var query string
	var args []interface{}

	if requestingUserRole == "admin" {
		// Admin can view any user
		query = `
			SELECT id, username, name, email, role, status, business_id, created_at, updated_at
			FROM users
			WHERE id = $1`
		args = []interface{}{id}
	} else {
		// Regular users can only view users from their business
		query = `
			SELECT id, username, name, email, role, status, business_id, created_at, updated_at
			FROM users
			WHERE id = $1 AND business_id = $2`
		args = []interface{}{id, businessID}
	}

	err = db.QueryRow(query, args...).Scan(
		&user.ID, &user.Username, &name, &user.Email, &user.Role, &user.Status, &user.BusinessID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Convert NULL name to empty string if needed
	if name.Valid {
		user.Name = name.String
	} else {
		user.Name = ""
	}

	log.Printf("GetUserByID called with id=%d, businessID=%d, role=%s", id, businessID, requestingUserRole)
	return user, nil
}

func (db *DB) GetUsers(businessID int) ([]models.User, error) {
	rows, err := db.Query(`
		SELECT id, username, name, email, role, status, business_id, created_at, updated_at
		FROM users
		WHERE business_id = $1
		ORDER BY created_at DESC`, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var name, status sql.NullString
		err := rows.Scan(&user.ID, &user.Username, &name, &user.Email, &user.Role, &status, &user.BusinessID, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if name.Valid {
			user.Name = name.String
		} else {
			user.Name = ""
		}
		if status.Valid {
			user.Status = status.String
		} else {
			user.Status = ""
		}
		users = append(users, user)
	}
	return users, nil
}

func (db *DB) GetUserStats() (*models.UserStats, error) {
	stats := &models.UserStats{}

	err := db.QueryRow(`
        SELECT 
            COUNT(*) as total,
            COUNT(CASE WHEN status = 'active' THEN 1 END) as active,
            COUNT(CASE WHEN status = 'inactive' THEN 1 END) as inactive,
            COUNT(CASE WHEN role = 'admin' THEN 1 END) as admins,
            COUNT(CASE WHEN created_at >= NOW() - INTERVAL '7 days' THEN 1 END) as new
        FROM users
    `).Scan(&stats.Total, &stats.Active, &stats.Inactive, &stats.Admins, &stats.New)

	return stats, err
}

func (db *DB) DeleteUser(id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

func (db *DB) UpdateUser(user *models.User) error {
	existingUser, err := db.GetUserByID(user.ID, user.BusinessID)
	if err != nil {
		log.Printf("UpdateUser: Error retrieving existing user data: %v", err)
		return err
	}

	if user.Username != "" {
		existingUser.Username = user.Username
	}
	if user.Name != "" {
		existingUser.Name = user.Name
	}
	if user.Email != "" {
		existingUser.Email = user.Email
	}
	if user.Role != "" {
		existingUser.Role = user.Role
	}
	if user.Status != "" {
		existingUser.Status = user.Status
	}

	_, err = db.Exec(`
        UPDATE users 
        SET username = $1, name = $2, email = $3, role = $4, status = $5, updated_at = $6
        WHERE id = $7`,
		existingUser.Username, existingUser.Name, existingUser.Email,
		existingUser.Role, existingUser.Status, time.Now(), user.ID,
	)

	if err != nil {
		log.Printf("UpdateUser: Error updating user ID %d: %v", user.ID, err)
	}

	return err
}

func (db *DB) CreateUser(user *models.User, businessID int) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Проверка на пустое имя - используем username если имя не задано
	if user.Name == "" {
		user.Name = user.Username
		log.Printf("CreateUser DB: Имя было пустым, используем username: %q", user.Name)
	}

	// Отладка: перед выполнением запроса
	log.Printf("CreateUser DB: Saving user - Username: %q, Name: %q, Email: %q, Role: %q, Status: %q, BusinessID: %d",
		user.Username, user.Name, user.Email, user.Role, user.Status, businessID)

	_, err = db.Exec(`
        INSERT INTO users (username, name, email, password, role, status, business_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)`,
		user.Username, user.Name, user.Email, string(hashedPassword), user.Role, user.Status, businessID, time.Now(),
	)

	if err != nil {
		log.Printf("CreateUser DB: Error creating user: %v", err)
	} else {
		log.Printf("CreateUser DB: User created successfully with business ID: %d", businessID)
	}

	return err
}

func (db *DB) GetStats() (map[string]int, error) {
	stats := map[string]int{}

	// Временные переменные для сканирования
	var total, active, inactive, new int

	// Общее количество пользователей
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// Активные пользователи
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE status = 'active'").Scan(&active)
	if err != nil {
		return nil, err
	}
	stats["active"] = active

	// Неактивные пользователи
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE status = 'inactive'").Scan(&inactive)
	if err != nil {
		return nil, err
	}
	stats["inactive"] = inactive

	// Новые пользователи (за последние 24 часа)
	err = db.QueryRow(`
        SELECT COUNT(*) FROM users 
        WHERE created_at >= NOW() - INTERVAL '24 hours'
    `).Scan(&new)
	if err != nil {
		return nil, err
	}
	stats["new"] = new

	return stats, nil
}

func (db *DB) GetAllInventory(businessID int) ([]models.Inventory, error) {
	rows, err := db.Query("SELECT id, name, category, quantity, unit, min_quantity, business_id, created_at, updated_at FROM inventory WHERE business_id = $1", businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Inventory
	for rows.Next() {
		var i models.Inventory
		err := rows.Scan(&i.ID, &i.Name, &i.Category, &i.Quantity, &i.Unit, &i.MinQuantity, &i.BusinessID, &i.CreatedAt, &i.UpdatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

func (db *DB) GetInventoryByID(id int, businessID int) (*models.Inventory, error) {
	var i models.Inventory
	err := db.QueryRow(`SELECT id, name, category, quantity, unit, min_quantity, business_id, created_at, updated_at FROM inventory WHERE id = $1 AND business_id = $2`, id, businessID).
		Scan(&i.ID, &i.Name, &i.Category, &i.Quantity, &i.Unit, &i.MinQuantity, &i.BusinessID, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("inventory not found")
		}
		return nil, err
	}
	return &i, nil
}

func (db *DB) CreateInventory(item *models.Inventory) error {
	err := db.QueryRow(`
		INSERT INTO inventory (name, category, quantity, unit, min_quantity, business_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id, created_at, updated_at`,
		item.Name, item.Category, item.Quantity, item.Unit, item.MinQuantity, item.BusinessID,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	return err
}

// UpdateInventory updates the quantity of an inventory item.
// It logs the change including the user who made the change (requires user_id in context).
func (db *DB) UpdateInventory(ctx context.Context, item *models.Inventory) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		log.Println("UpdateInventory: User ID not found in context for logging.")
	}

	var currentQuantity float64
	err := db.QueryRow("SELECT quantity FROM inventory WHERE id = $1 AND business_id = $2", item.ID, item.BusinessID).Scan(&currentQuantity)
	if err != nil {
		log.Printf("UpdateInventory: Failed to fetch current quantity for item %d: %v", item.ID, err)
	}

	_, err = db.Exec("UPDATE inventory SET name = $1, category = $2, quantity = $3, unit = $4, min_quantity = $5, updated_at = NOW() WHERE id = $6 AND business_id = $7", item.Name, item.Category, item.Quantity, item.Unit, item.MinQuantity, item.ID, item.BusinessID)
	if err != nil {
		log.Printf("Ошибка обновления запаса для ID %d: %v", item.ID, err)
		return fmt.Errorf("ошибка обновления запаса: %w", err)
	}

	log.Printf("Inventory update: Item ID %d, User ID %d (if available), Old Quantity %.2f, New Quantity %.2f", item.ID, userID, currentQuantity, item.Quantity)
	return nil
}

// DeleteInventory deletes an inventory item.
func (db *DB) DeleteInventory(id int, businessID int) error {
	_, err := db.Exec("DELETE FROM inventory WHERE id = $1 AND business_id = $2", id, businessID)
	return err
}

// Supplier methods
func (db *DB) GetAllSuppliers(businessID int) ([]models.Supplier, error) {
	query := `SELECT id, name, categories, phone, email, address, status, created_at, updated_at FROM suppliers WHERE business_id = $1`
	rows, err := db.Query(query, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []models.Supplier
	for rows.Next() {
		var s models.Supplier
		var categories []string
		if err := rows.Scan(&s.ID, &s.Name, pq.Array(&categories), &s.Phone, &s.Email, &s.Address, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.Categories = categories
		suppliers = append(suppliers, s)
	}
	return suppliers, nil
}

func (db *DB) GetSupplierByID(id int, businessID int) (*models.Supplier, error) {
	query := `SELECT id, name, categories, phone, email, address, status, created_at, updated_at FROM suppliers WHERE id = $1 AND business_id = $2`
	var s models.Supplier
	var categories []string
	err := db.QueryRow(query, id, businessID).Scan(&s.ID, &s.Name, pq.Array(&categories), &s.Phone, &s.Email, &s.Address, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	s.Categories = categories
	return &s, nil
}

func (db *DB) CreateSupplier(supplier *models.Supplier, businessID int) error {
	query := `
		INSERT INTO suppliers (name, categories, phone, email, address, status, business_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id`
	return db.QueryRow(
		query,
		supplier.Name,
		pq.Array(supplier.Categories),
		supplier.Phone,
		supplier.Email,
		supplier.Address,
		supplier.Status,
		businessID,
	).Scan(&supplier.ID)
}

func (db *DB) UpdateSupplier(supplier *models.Supplier, businessID int) error {
	query := `
		UPDATE suppliers
		SET name = $1, categories = $2, phone = $3, email = $4, address = $5, status = $6, updated_at = NOW()
		WHERE id = $7 AND business_id = $8`
	_, err := db.Exec(
		query,
		supplier.Name,
		pq.Array(supplier.Categories),
		supplier.Phone,
		supplier.Email,
		supplier.Address,
		supplier.Status,
		supplier.ID,
		businessID,
	)
	return err
}

func (db *DB) DeleteSupplier(id int, businessID int) error {
	query := `DELETE FROM suppliers WHERE id = $1 AND business_id = $2`
	_, err := db.Exec(query, id, businessID)
	return err
}

// Request methods
func (db *DB) GetAllRequests(businessID int) ([]models.Request, error) {
	query := `SELECT id, supplier_id, items, priority, comment, status, created_at, completed_at FROM requests WHERE business_id = $1`
	rows, err := db.Query(query, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.Request
	for rows.Next() {
		var r models.Request
		var items []string
		if err := rows.Scan(&r.ID, &r.SupplierID, pq.Array(&items), &r.Priority, &r.Comment, &r.Status, &r.CreatedAt, &r.CompletedAt); err != nil {
			return nil, err
		}
		r.Items = items
		requests = append(requests, r)
	}
	return requests, nil
}

func (db *DB) GetRequestByID(id int, businessID int) (*models.Request, error) {
	query := `SELECT id, supplier_id, items, priority, comment, status, created_at, completed_at FROM requests WHERE id = $1 AND business_id = $2`
	var r models.Request
	var items []string
	err := db.QueryRow(query, id, businessID).Scan(&r.ID, &r.SupplierID, pq.Array(&items), &r.Priority, &r.Comment, &r.Status, &r.CreatedAt, &r.CompletedAt)
	if err != nil {
		return nil, err
	}
	r.Items = items
	return &r, nil
}

func (db *DB) CreateRequest(request *models.Request, businessID int) error {
	query := `
		INSERT INTO requests (supplier_id, items, priority, comment, status, business_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id`
	return db.QueryRow(
		query,
		request.SupplierID,
		pq.Array(request.Items),
		request.Priority,
		request.Comment,
		request.Status,
		businessID,
	).Scan(&request.ID)
}

func (db *DB) UpdateRequest(request *models.Request, businessID int) error {
	query := `
		UPDATE requests
		SET status = $1
		WHERE id = $2 AND business_id = $3`
	_, err := db.Exec(
		query,
		request.Status,
		request.ID,
		businessID,
	)
	return err
}

func (db *DB) DeleteRequest(id int, businessID int) error {

	query := `DELETE FROM requests WHERE id = $1 AND business_id = $2`
	_, err := db.Exec(query, id, businessID)
	return err
}

// Table methods
func (db *DB) GetAllTables(businessID int) ([]models.Table, error) {
	rows, err := db.Query("SELECT id, number, seats, status, reserved_at, occupied_at FROM tables WHERE business_id = $1 OR business_id IS NULL ORDER BY number ASC", businessID)
	if err != nil {
		log.Printf("Error GetAllTables - querying tables: %v", err)
		return nil, err
	}
	defer rows.Close()

	var tables []models.Table
	for rows.Next() {
		var t models.Table
		if err := rows.Scan(&t.ID, &t.Number, &t.Seats, &t.Status, &t.ReservedAt, &t.OccupiedAt); err != nil {
			log.Printf("Error GetAllTables - scanning table row: %v", err)
			return nil, err
		}

		// Fetch active orders for this table
		orderRows, err := db.Query(`
            SELECT id, status, created_at, comment 
            FROM orders 
            WHERE table_id = $1 AND status NOT IN ('completed', 'cancelled') AND (business_id = $2 OR business_id IS NULL)
            ORDER BY created_at ASC`,
			t.ID, businessID,
		)
		if err != nil {
			log.Printf("Error GetAllTables - querying active orders for table %d: %v", t.ID, err)
			tables = append(tables, t)
			continue
		}
		defer orderRows.Close()

		var tableOrders []models.TableOrderInfo
		for orderRows.Next() {
			var toi models.TableOrderInfo
			var comment sql.NullString
			if err := orderRows.Scan(&toi.ID, &toi.Status, &toi.Time, &comment); err != nil {
				log.Printf("Error GetAllTables - scanning order row for table %d: %v", t.ID, err)
				continue
			}
			if comment.Valid {
				toi.Comment = &comment.String
			}
			tableOrders = append(tableOrders, toi)
		}
		orderRows.Close()

		if err := orderRows.Err(); err != nil {
			log.Printf("Error GetAllTables - iterating order rows for table %d: %v", t.ID, err)
		}

		t.Orders = tableOrders
		tables = append(tables, t)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error GetAllTables - iterating table rows: %v", err)
		return nil, err
	}

	return tables, nil
}

func (db *DB) GetTableByID(id int) (*models.Table, error) {
	query := `SELECT id, number, seats, status, reserved_at, occupied_at FROM tables WHERE id = $1`
	var t models.Table
	err := db.QueryRow(query, id).Scan(&t.ID, &t.Number, &t.Seats, &t.Status, &t.ReservedAt, &t.OccupiedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("table with ID %d not found", id)
		}
		log.Printf("Error GetTableByID - scanning table %d: %v", id, err)
		return nil, err
	}
	return &t, nil
}

func (db *DB) GetTableStats(businessID int) (*models.TableStats, error) {
	stats := &models.TableStats{}
	query := `
        SELECT
            COUNT(*) as total_tables,
            SUM(CASE WHEN status = 'free' THEN 1 ELSE 0 END) as free_tables,
            SUM(CASE WHEN status = 'occupied' THEN 1 ELSE 0 END) as occupied_tables,
            SUM(CASE WHEN status = 'reserved' THEN 1 ELSE 0 END) as reserved_tables,
            (SUM(CASE WHEN status = 'occupied' THEN 1 ELSE 0 END) * 100.0 / CASE WHEN COUNT(*) = 0 THEN 1 ELSE COUNT(*) END) as occupancy_percentage
        FROM tables
        WHERE business_id = $1
    `
	err := db.QueryRow(query, businessID).Scan(
		&stats.Total,
		&stats.Free,
		&stats.Occupied,
		&stats.Reserved,
		&stats.Occupancy,
	)
	if err != nil {
		log.Printf("Error GetTableStats query: %v", err)
		return nil, err
	}
	return stats, nil
}

func (db *DB) UpdateTableStatus(tableID int, status string) error {
	var occupiedAt sql.NullTime
	var reservedAt sql.NullTime

	switch models.TableStatus(status) { // Assuming models.TableStatus is defined for status constants
	case models.TableStatusOccupied:
		occupiedAt.Time = time.Now()
		occupiedAt.Valid = true
		reservedAt.Valid = false // Clear reservation if it becomes occupied
	case models.TableStatusReserved:
		reservedAt.Time = time.Now()
		reservedAt.Valid = true
		occupiedAt.Valid = false // Clear occupation if it becomes reserved (e.g. future reservation)
	case models.TableStatusFree:
		occupiedAt.Valid = false
		reservedAt.Valid = false
	default:
		// For any other status, don't explicitly change occupied_at or reserved_at
		// or handle as an error if status is unexpected
		log.Printf("Warning: UpdateTableStatus called with unhandled status '%s' for table %d. occupied_at and reserved_at will not be changed.", status, tableID)
		// Depending on strictness, you might want to return an error here or just update status and updated_at for unhandled cases.
		// For now, let's proceed to update only status and updated_at for unhandled cases.
		_, err := db.Exec("UPDATE tables SET status = $1 WHERE id = $2", status, tableID)
		if err != nil {
			log.Printf("Ошибка обновления статуса (без occupied_at/reserved_at) стола для ID %d: %v", tableID, err)
			return err
		}
		return nil
	}

	// updated_at is always set
	// Note: The original query used `updated_at = $2`. If `updated_at` is a specific column,
	// it should be explicitly part of the SET clause.
	// Assuming table `tables` also has an `updated_at` column that needs updating.
	result, err := db.Exec(`UPDATE tables 
						   SET status = $1, occupied_at = $2, reserved_at = $3 
						   WHERE id = $4`,
		status, occupiedAt, reservedAt, tableID)

	if err != nil {
		log.Printf("Ошибка обновления статуса стола для ID %d: %v", tableID, err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Ошибка получения количества затронутых строк для обновления статуса стола ID %d: %v", tableID, err)
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("стол с ID %d для обновления статуса не найден", tableID)
	}
	return nil
}

// GetDishByID retrieves a specific dish by its ID.
func (db *DB) GetDishByID(id int) (*models.Dish, error) {
	dish := &models.Dish{}
	err := db.QueryRow("SELECT id, name, category_id, price, is_available FROM dishes WHERE id = $1", id).
		Scan(&dish.ID, &dish.Name, &dish.CategoryID, &dish.Price, &dish.IsAvailable)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("dish with ID %d not found", id)
		}
		log.Printf("Error fetching dish by ID %d: %v", id, err)
		return nil, err
	}
	return dish, nil
}

// GetActiveOrdersWithItems retrieves all active orders along with their items.
func (db *DB) GetActiveOrdersWithItems(businessID int) ([]models.Order, error) {
	query := `
        SELECT o.id, o.table_id, o.waiter_id, o.status, o.comment, o.total_amount, 
               o.created_at, o.updated_at, o.completed_at, o.cancelled_at,
               COALESCE(
                   json_agg(
                       json_build_object(
                           'id', oi.id,
                           'dish_id', oi.dish_id,
                           'name', d.name,
                           'quantity', oi.quantity,
                           'price', oi.price,
                           'total', (oi.quantity * oi.price),
                           'notes', oi.notes
                       )
                   ) FILTER (WHERE oi.id IS NOT NULL), '[]'::json
               ) as items
        FROM orders o
        LEFT JOIN order_items oi ON o.id = oi.order_id
        LEFT JOIN dishes d ON oi.dish_id = d.id
        WHERE o.status IN ('new', 'accepted', 'preparing', 'ready', 'served')
        AND o.business_id = $1
        GROUP BY o.id
        ORDER BY o.created_at DESC
    `
	rows, err := db.Query(query, businessID)
	if err != nil {
		log.Printf("Error in GetActiveOrdersWithItems query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		var itemsJSON []byte
		var completedAt pq.NullTime
		var cancelledAt pq.NullTime

		err := rows.Scan(
			&order.ID, &order.TableID, &order.WaiterID, &order.Status, &order.Comment, &order.TotalAmount,
			&order.CreatedAt, &order.UpdatedAt, &completedAt, &cancelledAt, &itemsJSON,
		)
		if err != nil {
			log.Printf("Error scanning active order row: %v", err)
			return nil, err
		}

		if completedAt.Valid {
			order.CompletedAt = &completedAt.Time
		}
		if cancelledAt.Valid {
			order.CancelledAt = &cancelledAt.Time
		}

		if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
			log.Printf("Error unmarshalling order items for order %d: %v", order.ID, err)
			return nil, err
		}
		orders = append(orders, order)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error after iterating active order rows: %v", err)
		return nil, err
	}
	return orders, nil
}

// GetOrderByID retrieves a specific order by its ID, including its items.
func (db *DB) GetOrderByID(id int, businessID ...int) (*models.Order, error) {
	// If businessID is provided, use it for filtering
	whereClause := "WHERE o.id = $1"
	args := []interface{}{id}

	if len(businessID) > 0 && businessID[0] > 0 {
		whereClause += " AND o.business_id = $2"
		args = append(args, businessID[0])
	}

	query := `
        SELECT o.id, o.table_id, o.waiter_id, o.status, o.comment, o.total_amount, 
               o.created_at, o.updated_at, o.completed_at, o.cancelled_at,
               COALESCE(
                   json_agg(
                       json_build_object(
                           'id', oi.id,
						   'dish_id', oi.dish_id,
                           'name', d.name,
                           'quantity', oi.quantity,
                           'price', oi.price,
                           'total', (oi.quantity * oi.price),
						   'notes', oi.notes
                       )
                   ) FILTER (WHERE oi.id IS NOT NULL), '[]'::json
               ) as items
        FROM orders o
        LEFT JOIN order_items oi ON o.id = oi.order_id
        LEFT JOIN dishes d ON oi.dish_id = d.id
        ` + whereClause + `
        GROUP BY o.id`

	var o models.Order
	var itemsJSON []byte
	var completedAt, cancelledAt pq.NullTime

	err := db.QueryRow(query, args...).Scan(
		&o.ID, &o.TableID, &o.WaiterID, &o.Status, &o.Comment, &o.TotalAmount,
		&o.CreatedAt, &o.UpdatedAt, &completedAt, &cancelledAt, &itemsJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order with ID %d not found", id)
		}
		log.Printf("Error fetching order by ID %d: %v", id, err)
		return nil, err
	}

	if completedAt.Valid {
		o.CompletedAt = &completedAt.Time
	}
	if cancelledAt.Valid {
		o.CancelledAt = &cancelledAt.Time
	}

	if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
		log.Printf("Error unmarshalling items for order %d: %v", o.ID, err)
		return nil, err
	}
	return &o, nil
}

// CreateOrderAndItems creates a new order and its associated items in a transaction.
// It updates the order.ID and order.Items[i].ID upon successful creation.
func (db *DB) CreateOrderAndItems(order *models.Order, businessID int) (*models.Order, error) {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction for creating order: %v", err)
		return nil, err
	}

	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now

	orderSQL := `INSERT INTO orders (table_id, waiter_id, status, comment, total_amount, created_at, updated_at, business_id)
                 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`
	err = tx.QueryRow(orderSQL, order.TableID, order.WaiterID, order.Status, order.Comment, order.TotalAmount, order.CreatedAt, order.UpdatedAt, businessID).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		tx.Rollback()
		log.Printf("Error inserting order: %v", err)
		return nil, err
	}

	itemSQL := `INSERT INTO order_items (order_id, dish_id, quantity, price, notes, business_id)
                VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	for i := range order.Items {
		item := &order.Items[i]
		err = tx.QueryRow(itemSQL, order.ID, item.DishID, item.Quantity, item.Price, item.Notes, businessID).Scan(&item.ID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction for creating order: %v", err)
		return nil, err
	}
	return order, nil
}

// UpdateOrder updates an existing order's status and relevant timestamps.
// It expects order.Status, and potentially order.CompletedAt or order.CancelledAt to be set.
// order.ID must be valid.
func (db *DB) UpdateOrder(order *models.Order) error {
	query := `
        UPDATE orders 
        SET status = $1, comment = $2, total_amount = $3, 
            updated_at = $4, completed_at = $5, cancelled_at = $6
        WHERE id = $7`

	order.UpdatedAt = time.Now()

	_, err := db.Exec(query,
		order.Status, order.Comment, order.TotalAmount,
		order.UpdatedAt, order.CompletedAt, order.CancelledAt,
		order.ID,
	)
	if err != nil {
		log.Printf("Error updating order ID %d: %v", order.ID, err)
		return err
	}
	return nil
}

func (db *DB) GetOrderStatus(businessID int) (*models.OrderStats, error) {
	query := `
        SELECT 
            COUNT(CASE WHEN status NOT IN ('completed', 'cancelled') THEN 1 END) as total_active_orders,
            COUNT(CASE WHEN status = 'new' THEN 1 END) as new,
            COUNT(CASE WHEN status = 'accepted' THEN 1 END) as accepted,
            COUNT(CASE WHEN status = 'preparing' THEN 1 END) as preparing,
            COUNT(CASE WHEN status = 'ready' THEN 1 END) as ready,
            COUNT(CASE WHEN status = 'served' THEN 1 END) as served,
            COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_total,
            COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_total,
            COALESCE(SUM(CASE WHEN status = 'completed' THEN total_amount ELSE 0 END), 0) as completed_amount_total
        FROM orders
        WHERE (business_id = $1 OR business_id IS NULL)`

	var stats models.OrderStats
	err := db.QueryRow(query, businessID).Scan(
		&stats.TotalActiveOrders, &stats.New, &stats.Accepted, &stats.Preparing,
		&stats.Ready, &stats.Served, &stats.CompletedTotal, &stats.CancelledTotal,
		&stats.CompletedAmountTotal)
	if err != nil {
		log.Printf("Error fetching order stats: %v", err)
		return nil, err
	}
	return &stats, nil
}

// GetOrderHistoryWithItems retrieves completed or cancelled orders along with their items.
func (db *DB) GetOrderHistoryWithItems(businessID int) ([]models.Order, error) {
	query := `
    	SELECT o.id, o.table_id, o.waiter_id, o.status, o.comment, o.total_amount, 
       o.created_at, o.updated_at, o.completed_at, o.cancelled_at,
       COALESCE(
           json_agg(
               json_build_object(
                   'id', oi.id,
                   'dish_id', oi.dish_id,
                   'name', d.name,     
                   'quantity', oi.quantity,
                   'price', oi.price,
                   'total', oi.quantity * oi.price,
                   'notes', oi.notes
               )
           ) FILTER (WHERE oi.id IS NOT NULL), '[]'::json
       ) as items
		FROM orders o
		LEFT JOIN order_items oi ON o.id = oi.order_id
		LEFT JOIN dishes d ON oi.dish_id = d.id 
		WHERE o.status IN ('completed', 'cancelled')
		AND o.business_id = $1
		GROUP BY o.id 
		ORDER BY COALESCE(o.completed_at, o.cancelled_at, o.updated_at) DESC`

	rows, err := db.Query(query, businessID)
	if err != nil {
		log.Printf("Error in GetOrderHistoryWithItems query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		var itemsJSON []byte
		var completedAt, cancelledAt pq.NullTime

		err := rows.Scan(
			&o.ID, &o.TableID, &o.WaiterID, &o.Status, &o.Comment, &o.TotalAmount,
			&o.CreatedAt, &o.UpdatedAt, &completedAt, &cancelledAt, &itemsJSON,
		)
		if err != nil {
			log.Printf("Error scanning historical order: %v", err)
			return nil, err
		}

		if completedAt.Valid {
			o.CompletedAt = &completedAt.Time
		}
		if cancelledAt.Valid {
			o.CancelledAt = &cancelledAt.Time
		}

		if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
			log.Printf("Error unmarshalling items for historical order %d: %v", o.ID, err)
			return nil, err
		}
		orders = append(orders, o)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error after iterating historical order rows: %v", err)
		return nil, err
	}
	return orders, nil
}

// IsLastActiveOrderForTable checks if the given orderID is the last active order for the tableID.
func (db *DB) IsLastActiveOrderForTable(tableID, currentOrderID int) (bool, error) {
	var count int
	err := db.QueryRow(`
        SELECT COUNT(*) 
        FROM orders 
        WHERE table_id = $1 
          AND id != $2 
          AND status NOT IN ('completed', 'cancelled')`,
		tableID, currentOrderID,
	).Scan(&count)

	if err != nil {
		log.Printf("Error checking for other active orders for table %d (excluding order %d): %v", tableID, currentOrderID, err)
		return false, err
	}
	return count == 0, nil
}

// GetOrdersByStatus retrieves all orders with a specific status along with their items and dish categories.
func (db *DB) GetOrdersByStatus(status string, businessID int) ([]models.Order, error) {
	query := `
        SELECT o.id, o.table_id, o.waiter_id, o.status, o.comment, o.total_amount, 
               o.created_at, o.updated_at, o.completed_at, o.cancelled_at,
               COALESCE(
                   json_agg(
                       json_build_object(
                           'id', oi.id,
                           'dish_id', oi.dish_id,
                           'name', d.name,
                           'category', c.name,
                           'quantity', oi.quantity,
                           'price', oi.price,
                           'total', (oi.quantity * oi.price),
                           'notes', oi.notes
                       )
                   ) FILTER (WHERE oi.id IS NOT NULL), '[]'::json
               ) as items
        FROM orders o
        LEFT JOIN order_items oi ON o.id = oi.order_id
        LEFT JOIN dishes d ON oi.dish_id = d.id
        LEFT JOIN categories c ON d.category_id = c.id
        WHERE o.status = $1
        AND o.business_id = $2
        GROUP BY o.id
        ORDER BY o.created_at DESC
    `
	rows, err := db.Query(query, status, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		var itemsJSON []byte
		var completedAt, cancelledAt sql.NullTime

		err := rows.Scan(
			&order.ID, &order.TableID, &order.WaiterID, &order.Status, &order.Comment, &order.TotalAmount,
			&order.CreatedAt, &order.UpdatedAt, &completedAt, &cancelledAt, &itemsJSON,
		)
		if err != nil {
			return nil, err
		}
		if completedAt.Valid {
			order.CompletedAt = &completedAt.Time
		}
		if cancelledAt.Valid {
			order.CancelledAt = &cancelledAt.Time
		}
		if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// TableHasActiveOrders checks if a table has any active orders
func (db *DB) TableHasActiveOrders(tableID int) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM orders 
		WHERE table_id = $1 
		AND status NOT IN ('completed', 'cancelled')`,
		tableID,
	).Scan(&count)

	if err != nil {
		log.Printf("Error checking for active orders for table %d: %v", tableID, err)
		return false, err
	}
	return count > 0, nil
}

// UpdateTableStatusWithTimes updates a table's status and timestamp fields
func (db *DB) UpdateTableStatusWithTimes(tableID int, status string, reservedAt, occupiedAt *time.Time) error {
	query := `
		UPDATE tables 
		SET status = $1, reserved_at = $2, occupied_at = $3
		WHERE id = $4`

	result, err := db.Exec(query, status, reservedAt, occupiedAt, tableID)
	if err != nil {
		log.Printf("Error updating table %d status and timestamps: %v", tableID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting affected rows for table %d status update: %v", tableID, err)
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("table with ID %d not found", tableID)
	}

	return nil
}

// GetUserBusinessID retrieves the business ID associated with a user
func (db *DB) GetUserBusinessID(userID int) (int, error) {
	var businessID int

	// First check the users table for a direct association
	err := db.QueryRow(`
        SELECT business_id FROM users WHERE id = $1
    `, userID).Scan(&businessID)

	if err == nil && businessID > 0 {
		log.Printf("Found business ID %d for user %d in users table", businessID, userID)
		return businessID, nil
	}

	// If not found in users table, check user_businesses table if it exists
	// This is for systems with many-to-many user-business relationships
	var exists bool
	err = db.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM information_schema.tables 
            WHERE table_name = 'user_businesses'
        )
    `).Scan(&exists)

	if err != nil {
		log.Printf("Error checking if user_businesses table exists: %v", err)
		return 0, err
	}

	if exists {
		// Check for the user's primary business in user_businesses table
		err = db.QueryRow(`
            SELECT business_id FROM user_businesses 
            WHERE user_id = $1 AND is_primary = true
        `, userID).Scan(&businessID)

		if err == nil && businessID > 0 {
			log.Printf("Found primary business ID %d for user %d in user_businesses table", businessID, userID)
			return businessID, nil
		}

		// If no primary business found, get the first business associated with the user
		err = db.QueryRow(`
            SELECT business_id FROM user_businesses 
            WHERE user_id = $1 
            LIMIT 1
        `, userID).Scan(&businessID)

		if err == nil && businessID > 0 {
			log.Printf("Found business ID %d for user %d in user_businesses table", businessID, userID)
			return businessID, nil
		}
	}

	// If no business found or there was an error
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error retrieving business ID for user %d: %v", userID, err)
		return 0, err
	}

	log.Printf("No business found for user %d", userID)
	return 0, nil
}
