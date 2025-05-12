package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"restaurant-management/configs"
	"restaurant-management/internal/models"
	"strconv"
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
		if err == sql.ErrNoRows {
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

func (db *DB) GetUsers(page, limit int, status, role, search string) ([]models.User, int, error) {
	// Базовый запрос
	query := `
        SELECT id, username, email, role, status, last_active
        FROM users
        WHERE 1=1
    `
	countQuery := `SELECT COUNT(*) FROM users WHERE 1=1`
	args := []interface{}{}

	// Добавляем фильтры
	if status != "" {
		query += ` AND status = $` + strconv.Itoa(len(args)+1)
		countQuery += ` AND status = $` + strconv.Itoa(len(args)+1)
		args = append(args, status)
	}

	if role != "" {
		query += ` AND role = $` + strconv.Itoa(len(args)+1)
		countQuery += ` AND role = $` + strconv.Itoa(len(args)+1)
		args = append(args, role)
	}

	if search != "" {
		query += ` AND (username ILIKE $` + strconv.Itoa(len(args)+1) +
			` OR email ILIKE $` + strconv.Itoa(len(args)+1) + `)`
		countQuery += ` AND (username ILIKE $` + strconv.Itoa(len(args)+1) +
			` OR email ILIKE $` + strconv.Itoa(len(args)+1) + `)`
		args = append(args, "%"+search+"%")
	}

	// Получаем общее количество
	var total int
	err := db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Добавляем пагинацию
	query += ` ORDER BY id DESC LIMIT $` + strconv.Itoa(len(args)+1) +
		` OFFSET $` + strconv.Itoa(len(args)+2)
	args = append(args, limit, (page-1)*limit)

	// Выполняем запрос
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var lastActive sql.NullTime
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.Status, &lastActive)
		if err != nil {
			return nil, 0, err
		}
		if lastActive.Valid {
			u.LastActive = lastActive.Time
		}
		users = append(users, u)
	}

	return users, total, nil
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
	_, err := db.Exec(`
        UPDATE users 
        SET username = $1, email = $2, role = $3, status = $4, updated_at = $5
        WHERE id = $6`,
		user.Username, user.Email, user.Role, user.Status, time.Now(), user.ID,
	)
	return err
}

func (db *DB) CreateUser(user *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
        INSERT INTO users (username, email, password, role, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $6)`,
		user.Username, user.Email, string(hashedPassword), user.Role, user.Status, time.Now(),
	)
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

func (db *DB) GetAllInventory() ([]models.Inventory, error) {
	rows, err := db.Query("SELECT id, name, category, quantity, unit, min_quantity, branch, created_at, updated_at FROM inventory")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Inventory
	for rows.Next() {
		var i models.Inventory
		err := rows.Scan(&i.ID, &i.Name, &i.Category, &i.Quantity, &i.Unit, &i.MinQuantity, &i.Branch, &i.CreatedAt, &i.UpdatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

func (db *DB) GetInventoryByID(id int) (*models.Inventory, error) {
	var i models.Inventory
	err := db.QueryRow(`SELECT id, name, category, quantity, unit, min_quantity, branch, created_at, updated_at FROM inventory WHERE id = $1`, id).
		Scan(&i.ID, &i.Name, &i.Category, &i.Quantity, &i.Unit, &i.MinQuantity, &i.Branch, &i.CreatedAt, &i.UpdatedAt)
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
		INSERT INTO inventory (name, category, quantity, unit, min_quantity, branch, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id, created_at, updated_at`,
		item.Name, item.Category, item.Quantity, item.Unit, item.MinQuantity, item.Branch,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	return err
}

func (db *DB) UpdateInventory(item *models.Inventory) error {
	_, err := db.Exec(`
		UPDATE inventory SET name = $1, category = $2, quantity = $3, unit = $4, min_quantity = $5, branch = $6, updated_at = NOW()
		WHERE id = $7`,
		item.Name, item.Category, item.Quantity, item.Unit, item.MinQuantity, item.Branch, item.ID,
	)
	return err
}

func (db *DB) DeleteInventory(id int) error {
	_, err := db.Exec("DELETE FROM inventory WHERE id = $1", id)
	return err
}

// Supplier methods
func (db *DB) GetAllSuppliers() ([]models.Supplier, error) {
	query := `SELECT id, name, categories, phone, email, address, status, created_at, updated_at FROM suppliers`
	rows, err := db.Query(query)
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

func (db *DB) GetSupplierByID(id int) (*models.Supplier, error) {
	query := `SELECT id, name, categories, phone, email, address, status, created_at, updated_at FROM suppliers WHERE id = $1`
	var s models.Supplier
	var categories []string
	err := db.QueryRow(query, id).Scan(&s.ID, &s.Name, pq.Array(&categories), &s.Phone, &s.Email, &s.Address, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	s.Categories = categories
	return &s, nil
}

func (db *DB) CreateSupplier(supplier *models.Supplier) error {
	query := `
		INSERT INTO suppliers (name, categories, phone, email, address, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id`
	return db.QueryRow(
		query,
		supplier.Name,
		pq.Array(supplier.Categories),
		supplier.Phone,
		supplier.Email,
		supplier.Address,
		supplier.Status,
	).Scan(&supplier.ID)
}

func (db *DB) UpdateSupplier(supplier *models.Supplier) error {
	query := `
		UPDATE suppliers
		SET name = $1, categories = $2, phone = $3, email = $4, address = $5, status = $6, updated_at = NOW()
		WHERE id = $7`
	_, err := db.Exec(
		query,
		supplier.Name,
		pq.Array(supplier.Categories),
		supplier.Phone,
		supplier.Email,
		supplier.Address,
		supplier.Status,
		supplier.ID,
	)
	return err
}

func (db *DB) DeleteSupplier(id int) error {
	query := `DELETE FROM suppliers WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

// Request methods
func (db *DB) GetAllRequests() ([]models.Request, error) {
	query := `SELECT id, branch, supplier_id, items, priority, comment, status, created_at, completed_at FROM requests`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.Request
	for rows.Next() {
		var r models.Request
		var items []string
		if err := rows.Scan(&r.ID, &r.Branch, &r.SupplierID, pq.Array(&items), &r.Priority, &r.Comment, &r.Status, &r.CreatedAt, &r.CompletedAt); err != nil {
			return nil, err
		}
		r.Items = items
		requests = append(requests, r)
	}
	return requests, nil
}

func (db *DB) GetRequestByID(id int) (*models.Request, error) {
	query := `SELECT id, branch, supplier_id, items, priority, comment, status, created_at, completed_at FROM requests WHERE id = $1`
	var r models.Request
	var items []string
	err := db.QueryRow(query, id).Scan(&r.ID, &r.Branch, &r.SupplierID, pq.Array(&items), &r.Priority, &r.Comment, &r.Status, &r.CreatedAt, &r.CompletedAt)
	if err != nil {
		return nil, err
	}
	r.Items = items
	return &r, nil
}

func (db *DB) CreateRequest(request *models.Request) error {
	query := `
		INSERT INTO requests (branch, supplier_id, items, priority, comment, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id`
	return db.QueryRow(
		query,
		request.Branch,
		request.SupplierID,
		pq.Array(request.Items),
		request.Priority,
		request.Comment,
		request.Status,
	).Scan(&request.ID)
}

func (db *DB) UpdateRequest(request *models.Request) error {
	query := `
		UPDATE requests
		SET branch = $1, supplier_id = $2, items = $3, priority = $4, comment = $5, status = $6, completed_at = $7
		WHERE id = $8`
	_, err := db.Exec(
		query,
		request.Branch,
		request.SupplierID,
		pq.Array(request.Items),
		request.Priority,
		request.Comment,
		request.Status,
		request.CompletedAt,
		request.ID,
	)
	return err
}

func (db *DB) DeleteRequest(id int) error {
	query := `DELETE FROM requests WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

// Table methods
func (db *DB) GetAllTables() ([]models.Table, error) {
	query := `SELECT id, number, seats, status, created_at, updated_at FROM tables`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []models.Table
	for rows.Next() {
		var t models.Table
		if err := rows.Scan(&t.ID, &t.Number, &t.Seats, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, nil
}

func (db *DB) GetTableByID(id int) (*models.Table, error) {
	query := `SELECT id, number, seats, status, created_at, updated_at FROM tables WHERE id = $1`
	var t models.Table
	err := db.QueryRow(query, id).Scan(&t.ID, &t.Number, &t.Seats, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (db *DB) GetTableStatus() (*models.TableStatus, error) {
	query := `
        SELECT 
            COUNT(*) as total,
            COUNT(CASE WHEN status = 'free' THEN 1 END) as free,
            COUNT(CASE WHEN status = 'occupied' THEN 1 END) as occupied,
            COUNT(CASE WHEN status = 'reserved' THEN 1 END) as reserved
        FROM tables`
	var status models.TableStatus
	err := db.QueryRow(query).Scan(&status.Total, &status.Free, &status.Occupied, &status.Reserved)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (db *DB) UpdateTableStatus(table *models.Table) error {
	query := `
        UPDATE tables 
        SET status = $1, updated_at = NOW()
        WHERE id = $2`
	_, err := db.Exec(query, table.Status, table.ID)
	return err
}

// Order methods
func (db *DB) GetAllOrders() ([]models.Order, error) {
	query := `
        SELECT o.id, o.table_id, o.waiter_id, o.status, o.comment, o.total, 
               o.created_at, o.updated_at, o.completed_at, o.cancelled_at,
               json_agg(json_build_object(
                   'id', oi.id,
                   'name', oi.name,
                   'quantity', oi.quantity,
                   'price', oi.price,
                   'total', oi.total
               )) as items
        FROM orders o
        LEFT JOIN order_items oi ON o.id = oi.order_id
        WHERE o.status NOT IN ('completed', 'cancelled')
        GROUP BY o.id
        ORDER BY o.created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		var items []byte
		if err := rows.Scan(&o.ID, &o.TableID, &o.WaiterID, &o.Status, &o.Comment, &o.Total,
			&o.CreatedAt, &o.UpdatedAt, &o.CompletedAt, &o.CancelledAt, &items); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(items, &o.Items); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (db *DB) GetOrderByID(id int) (*models.Order, error) {
	query := `
        SELECT o.id, o.table_id, o.waiter_id, o.status, o.comment, o.total, 
               o.created_at, o.updated_at, o.completed_at, o.cancelled_at,
               json_agg(json_build_object(
                   'id', oi.id,
                   'name', oi.name,
                   'quantity', oi.quantity,
                   'price', oi.price,
                   'total', oi.total
               )) as items
        FROM orders o
        LEFT JOIN order_items oi ON o.id = oi.order_id
        WHERE o.id = $1
        GROUP BY o.id`

	var o models.Order
	var items []byte
	err := db.QueryRow(query, id).Scan(&o.ID, &o.TableID, &o.WaiterID, &o.Status, &o.Comment, &o.Total,
		&o.CreatedAt, &o.UpdatedAt, &o.CompletedAt, &o.CancelledAt, &items)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(items, &o.Items); err != nil {
		return nil, err
	}
	return &o, nil
}

func (db *DB) CreateOrder(order *models.Order) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert order
	query := `
        INSERT INTO orders (table_id, waiter_id, status, comment, total, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
        RETURNING id`
	err = tx.QueryRow(query, order.TableID, order.WaiterID, order.Status, order.Comment, order.Total).Scan(&order.ID)
	if err != nil {
		return err
	}

	// Insert order items
	for _, item := range order.Items {
		_, err = tx.Exec(`
            INSERT INTO order_items (order_id, name, quantity, price, total)
            VALUES ($1, $2, $3, $4, $5)`,
			order.ID, item.Name, item.Quantity, item.Price, item.Total)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) UpdateOrder(order *models.Order) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update order
	query := `
        UPDATE orders 
        SET status = $1, comment = $2, total = $3, updated_at = NOW(),
            completed_at = CASE WHEN $1 = 'completed' THEN NOW() ELSE completed_at END,
            cancelled_at = CASE WHEN $1 = 'cancelled' THEN NOW() ELSE cancelled_at END
        WHERE id = $4`
	_, err = tx.Exec(query, order.Status, order.Comment, order.Total, order.ID)
	if err != nil {
		return err
	}

	// Update order items if provided
	if len(order.Items) > 0 {
		// Delete existing items
		_, err = tx.Exec("DELETE FROM order_items WHERE order_id = $1", order.ID)
		if err != nil {
			return err
		}

		// Insert new items
		for _, item := range order.Items {
			_, err = tx.Exec(`
                INSERT INTO order_items (order_id, name, quantity, price, total)
                VALUES ($1, $2, $3, $4, $5)`,
				order.ID, item.Name, item.Quantity, item.Price, item.Total)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (db *DB) GetOrderStatus() (*models.OrderStatus, error) {
	query := `
        SELECT 
            COUNT(*) as total,
            COUNT(CASE WHEN status = 'new' THEN 1 END) as new,
            COUNT(CASE WHEN status = 'accepted' THEN 1 END) as accepted,
            COUNT(CASE WHEN status = 'preparing' THEN 1 END) as preparing,
            COUNT(CASE WHEN status = 'ready' THEN 1 END) as ready,
            COUNT(CASE WHEN status = 'served' THEN 1 END) as served,
            COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed,
            COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled,
            COALESCE(SUM(CASE WHEN status = 'completed' THEN total ELSE 0 END), 0) as total_amount
        FROM orders`

	var status models.OrderStatus
	err := db.QueryRow(query).Scan(
		&status.Total, &status.New, &status.Accepted, &status.Preparing,
		&status.Ready, &status.Served, &status.Completed, &status.Cancelled,
		&status.TotalAmount)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (db *DB) GetOrderHistory() ([]models.Order, error) {
	query := `
        SELECT o.id, o.table_id, o.waiter_id, o.status, o.comment, o.total, 
               o.created_at, o.updated_at, o.completed_at, o.cancelled_at,
               json_agg(json_build_object(
                   'id', oi.id,
                   'name', oi.name,
                   'quantity', oi.quantity,
                   'price', oi.price,
                   'total', oi.total
               )) as items
        FROM orders o
        LEFT JOIN order_items oi ON o.id = oi.order_id
        WHERE o.status IN ('completed', 'cancelled')
        GROUP BY o.id
        ORDER BY COALESCE(o.completed_at, o.cancelled_at) DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		var items []byte
		if err := rows.Scan(&o.ID, &o.TableID, &o.WaiterID, &o.Status, &o.Comment, &o.Total,
			&o.CreatedAt, &o.UpdatedAt, &o.CompletedAt, &o.CancelledAt, &items); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(items, &o.Items); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}
