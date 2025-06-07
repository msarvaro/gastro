package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"restaurant-management/internal/domain/consts" // Added import
	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
)

// orderRepository implements the OrderRepository interface using PostgreSQL
type orderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *sql.DB) repository.OrderRepository {
	return &orderRepository{
		db: db,
	}
}

// GetByID retrieves an order by ID
func (r *orderRepository) GetByID(ctx context.Context, id int) (*entity.Order, error) {
	query := `
		SELECT id, table_id, waiter_id, shift_id, status, total_amount, comment,
		       created_at, updated_at, completed_at, cancelled_at, business_id
		FROM orders
		WHERE id = $1
	`

	var order entity.Order
	var shiftID sql.NullInt64
	var completedAt, cancelledAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.TableID,
		&order.WaiterID,
		&shiftID,
		&order.Status,
		&order.TotalAmount,
		&order.Comment,
		&order.CreatedAt,
		&order.UpdatedAt,
		&completedAt,
		&cancelledAt,
		&order.BusinessID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if shiftID.Valid {
		sid := int(shiftID.Int64)
		order.ShiftID = &sid
	}

	if completedAt.Valid {
		order.CompletedAt = &completedAt.Time
	}

	if cancelledAt.Valid {
		order.CancelledAt = &cancelledAt.Time
	}

	// Load order items
	items, err := r.GetOrderItems(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	order.Items = items

	return &order, nil
}

// GetByBusinessID retrieves orders for a business with pagination
func (r *orderRepository) GetByBusinessID(ctx context.Context, businessID, limit, offset int) ([]*entity.Order, error) {
	query := `
		SELECT id, table_id, waiter_id, shift_id, status, total_amount, comment,
		       created_at, updated_at, completed_at, cancelled_at, business_id
		FROM orders
		WHERE business_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, businessID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*entity.Order

	for rows.Next() {
		var order entity.Order
		var shiftID sql.NullInt64
		var completedAt, cancelledAt sql.NullTime

		err := rows.Scan(
			&order.ID,
			&order.TableID,
			&order.WaiterID,
			&shiftID,
			&order.Status,
			&order.TotalAmount,
			&order.Comment,
			&order.CreatedAt,
			&order.UpdatedAt,
			&completedAt,
			&cancelledAt,
			&order.BusinessID,
		)

		if err != nil {
			return nil, err
		}

		if shiftID.Valid {
			sid := int(shiftID.Int64)
			order.ShiftID = &sid
		}

		if completedAt.Valid {
			order.CompletedAt = &completedAt.Time
		}

		if cancelledAt.Valid {
			order.CancelledAt = &cancelledAt.Time
		}

		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Load items for each order
	for _, order := range orders {
		items, err := r.GetOrderItems(ctx, order.ID)
		if err != nil {
			return nil, err
		}

		order.Items = items
	}

	return orders, nil
}

// Create adds a new order
func (r *orderRepository) Create(ctx context.Context, order *entity.Order) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Insert order
	query := `
		INSERT INTO orders (table_id, waiter_id, shift_id, status, total_amount, comment,
		                   created_at, updated_at, business_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var shiftID sql.NullInt64
	if order.ShiftID != nil {
		shiftID.Int64 = int64(*order.ShiftID)
		shiftID.Valid = true
	}

	now := time.Now()
	if order.CreatedAt.IsZero() {
		order.CreatedAt = now
	}
	if order.UpdatedAt.IsZero() {
		order.UpdatedAt = now
	}

	err = tx.QueryRowContext(
		ctx,
		query,
		order.TableID,
		order.WaiterID,
		shiftID,
		order.Status,
		order.TotalAmount,
		order.Comment,
		order.CreatedAt,
		order.UpdatedAt,
		order.BusinessID,
	).Scan(&order.ID)

	if err != nil {
		return err
	}

	// Insert order items
	for _, item := range order.Items {
		item.OrderID = order.ID
		item.BusinessID = order.BusinessID

		err = r.AddOrderItem(ctx, order.ID, item)
		if err != nil {
			return err
		}
	}

	// Update table status to occupied if it's not already
	tableQuery := `
		UPDATE tables
		SET status = $1, occupied_at = $2
		WHERE id = $3 AND status != $1
	`

	_, err = tx.ExecContext(ctx, tableQuery, consts.TableStatusOccupied, now, order.TableID)
	if err != nil {
		return err
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Update updates an existing order
func (r *orderRepository) Update(ctx context.Context, order *entity.Order) error {
	query := `
		UPDATE orders
		SET table_id = $1, waiter_id = $2, shift_id = $3, status = $4,
		    total_amount = $5, comment = $6, updated_at = $7,
		    completed_at = $8, cancelled_at = $9
		WHERE id = $10 AND business_id = $11
	`

	var shiftID sql.NullInt64
	if order.ShiftID != nil {
		shiftID.Int64 = int64(*order.ShiftID)
		shiftID.Valid = true
	}

	var completedAt, cancelledAt sql.NullTime
	if order.CompletedAt != nil {
		completedAt.Time = *order.CompletedAt
		completedAt.Valid = true
	}

	if order.CancelledAt != nil {
		cancelledAt.Time = *order.CancelledAt
		cancelledAt.Valid = true
	}

	order.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		order.TableID,
		order.WaiterID,
		shiftID,
		order.Status,
		order.TotalAmount,
		order.Comment,
		order.UpdatedAt,
		completedAt,
		cancelledAt,
		order.ID,
		order.BusinessID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("order not found")
	}

	return nil
}

// Delete removes an order
func (r *orderRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM orders WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("order not found")
	}

	return nil
}

// UpdateStatus updates an order's status
func (r *orderRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = $2
	`

	params := []interface{}{status, time.Now()}
	paramCount := 3 // Starting parameter count

	// Add timestamps based on status
	if status == consts.OrderStatusPaid { // Changed "completed"
		query += ", completed_at = $" + strconv.Itoa(paramCount)
		params = append(params, time.Now())
		paramCount++
	} else if status == consts.OrderStatusCanceled { // Changed "cancelled"
		query += ", cancelled_at = $" + strconv.Itoa(paramCount)
		params = append(params, time.Now())
		paramCount++
	}

	query += " WHERE id = $" + strconv.Itoa(paramCount)
	params = append(params, id)

	result, err := r.db.ExecContext(ctx, query, params...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("order not found")
	}

	return nil
}

// GetByStatus retrieves orders by status
func (r *orderRepository) GetByStatus(ctx context.Context, businessID int, status string) ([]*entity.Order, error) {
	query := `
		SELECT id, table_id, waiter_id, shift_id, status, total_amount, comment,
		       created_at, updated_at, completed_at, cancelled_at, business_id
		FROM orders
		WHERE business_id = $1 AND status = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, businessID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*entity.Order

	for rows.Next() {
		var order entity.Order
		var shiftID sql.NullInt64
		var completedAt, cancelledAt sql.NullTime

		err := rows.Scan(
			&order.ID,
			&order.TableID,
			&order.WaiterID,
			&shiftID,
			&order.Status,
			&order.TotalAmount,
			&order.Comment,
			&order.CreatedAt,
			&order.UpdatedAt,
			&completedAt,
			&cancelledAt,
			&order.BusinessID,
		)

		if err != nil {
			return nil, err
		}

		if shiftID.Valid {
			sid := int(shiftID.Int64)
			order.ShiftID = &sid
		}

		if completedAt.Valid {
			order.CompletedAt = &completedAt.Time
		}

		if cancelledAt.Valid {
			order.CancelledAt = &cancelledAt.Time
		}

		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Load items for each order
	for _, order := range orders {
		items, err := r.GetOrderItems(ctx, order.ID)
		if err != nil {
			return nil, err
		}

		order.Items = items
	}

	return orders, nil
}

// GetByWaiterID retrieves orders for a waiter
func (r *orderRepository) GetByWaiterID(ctx context.Context, waiterID int) ([]*entity.Order, error) {
	query := `
		SELECT id, table_id, waiter_id, shift_id, status, total_amount, comment,
		       created_at, updated_at, completed_at, cancelled_at, business_id
		FROM orders
		WHERE waiter_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, waiterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*entity.Order

	for rows.Next() {
		var order entity.Order
		var shiftID sql.NullInt64
		var completedAt, cancelledAt sql.NullTime

		err := rows.Scan(
			&order.ID,
			&order.TableID,
			&order.WaiterID,
			&shiftID,
			&order.Status,
			&order.TotalAmount,
			&order.Comment,
			&order.CreatedAt,
			&order.UpdatedAt,
			&completedAt,
			&cancelledAt,
			&order.BusinessID,
		)

		if err != nil {
			return nil, err
		}

		if shiftID.Valid {
			sid := int(shiftID.Int64)
			order.ShiftID = &sid
		}

		if completedAt.Valid {
			order.CompletedAt = &completedAt.Time
		}

		if cancelledAt.Valid {
			order.CancelledAt = &cancelledAt.Time
		}

		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// GetWaiterOrderStatistics retrieves order statistics for a waiter
func (r *orderRepository) GetWaiterOrderStatistics(ctx context.Context, waiterID int) (map[string]int, error) {
	query := `
		SELECT
			COUNT(*) FILTER (WHERE status = $3) AS new,
			COUNT(*) FILTER (WHERE status = $4) AS accepted,
			COUNT(*) FILTER (WHERE status = $5) AS preparing,
			COUNT(*) FILTER (WHERE status = $6) AS ready,
			COUNT(*) FILTER (WHERE status = $7) AS served,
			COUNT(*) FILTER (WHERE status = $8) AS completed,
			COUNT(*) FILTER (WHERE status = $9) AS cancelled,
			COUNT(*) AS total
		FROM orders
		WHERE waiter_id = $1 AND (completed_at IS NULL OR completed_at > $2)
	`

	// Consider orders from the last 24 hours
	oneDayAgo := time.Now().Add(-24 * time.Hour)

	var new, accepted, preparing, ready, served, completed, cancelled, total int

	err := r.db.QueryRowContext(ctx, query, waiterID, oneDayAgo,
		consts.OrderStatusPending,   // $3 new
		consts.OrderStatusConfirmed, // $4 accepted
		consts.OrderStatusPreparing, // $5 preparing
		consts.OrderStatusReady,     // $6 ready
		consts.OrderStatusDelivered, // $7 served
		consts.OrderStatusPaid,      // $8 completed
		consts.OrderStatusCanceled,  // $9 cancelled
	).Scan(
		&new,
		&accepted,
		&preparing,
		&ready,
		&served,
		&completed,
		&cancelled,
		&total,
	)

	if err != nil {
		return nil, err
	}

	stats := map[string]int{
		"new":       new,
		"accepted":  accepted,
		"preparing": preparing,
		"ready":     ready,
		"served":    served,
		"completed": completed,
		"cancelled": cancelled,
		"total":     total,
	}

	return stats, nil
}

// GetWaiterTablesServedCount retrieves the count of tables served by a waiter
func (r *orderRepository) GetWaiterTablesServedCount(ctx context.Context, waiterID int) (int, error) {
	query := `
		SELECT COUNT(DISTINCT table_id)
		FROM orders
		WHERE waiter_id = $1 AND status IN ($2, $3)
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, waiterID, consts.OrderStatusPaid, consts.OrderStatusDelivered).Scan(&count) // Changed
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetWaiterCompletedOrdersCount retrieves the count of completed orders by a waiter
func (r *orderRepository) GetWaiterCompletedOrdersCount(ctx context.Context, waiterID int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM orders
		WHERE waiter_id = $1 AND status = $2
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, waiterID, consts.OrderStatusPaid).Scan(&count) // Changed
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetByDateRange retrieves orders within a date range
func (r *orderRepository) GetByDateRange(ctx context.Context, businessID int, startDate, endDate time.Time) ([]*entity.Order, error) {
	query := `
		SELECT id, table_id, waiter_id, shift_id, status, total_amount, comment,
		       created_at, updated_at, completed_at, cancelled_at, business_id
		FROM orders
		WHERE business_id = $1 AND created_at BETWEEN $2 AND $3
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, businessID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*entity.Order

	for rows.Next() {
		var order entity.Order
		var shiftID sql.NullInt64
		var completedAt, cancelledAt sql.NullTime

		err := rows.Scan(
			&order.ID,
			&order.TableID,
			&order.WaiterID,
			&shiftID,
			&order.Status,
			&order.TotalAmount,
			&order.Comment,
			&order.CreatedAt,
			&order.UpdatedAt,
			&completedAt,
			&cancelledAt,
			&order.BusinessID,
		)

		if err != nil {
			return nil, err
		}

		if shiftID.Valid {
			sid := int(shiftID.Int64)
			order.ShiftID = &sid
		}

		if completedAt.Valid {
			order.CompletedAt = &completedAt.Time
		}

		if cancelledAt.Valid {
			order.CancelledAt = &cancelledAt.Time
		}

		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// GetByTableID retrieves orders for a table
func (r *orderRepository) GetByTableID(ctx context.Context, tableID int) ([]*entity.Order, error) {
	query := `
		SELECT id, table_id, waiter_id, shift_id, status, total_amount, comment,
		       created_at, updated_at, completed_at, cancelled_at, business_id
		FROM orders
		WHERE table_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tableID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*entity.Order

	for rows.Next() {
		var order entity.Order
		var shiftID sql.NullInt64
		var completedAt, cancelledAt sql.NullTime

		err := rows.Scan(
			&order.ID,
			&order.TableID,
			&order.WaiterID,
			&shiftID,
			&order.Status,
			&order.TotalAmount,
			&order.Comment,
			&order.CreatedAt,
			&order.UpdatedAt,
			&completedAt,
			&cancelledAt,
			&order.BusinessID,
		)

		if err != nil {
			return nil, err
		}

		if shiftID.Valid {
			sid := int(shiftID.Int64)
			order.ShiftID = &sid
		}

		if completedAt.Valid {
			order.CompletedAt = &completedAt.Time
		}

		if cancelledAt.Valid {
			order.CancelledAt = &cancelledAt.Time
		}

		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// GetActiveByTableID retrieves the active order for a table
func (r *orderRepository) GetActiveByTableID(ctx context.Context, tableID int) (*entity.Order, error) {
	query := `
		SELECT id, table_id, waiter_id, shift_id, status, total_amount, comment,
		       created_at, updated_at, completed_at, cancelled_at, business_id
		FROM orders
		WHERE table_id = $1 AND status NOT IN ($2, $3)
		ORDER BY created_at DESC
		LIMIT 1
	`

	var order entity.Order
	var shiftID sql.NullInt64
	var completedAt, cancelledAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, tableID, consts.OrderStatusPaid, consts.OrderStatusCanceled).Scan( // Changed
		&order.ID,
		&order.TableID,
		&order.WaiterID,
		&shiftID,
		&order.Status,
		&order.TotalAmount,
		&order.Comment,
		&order.CreatedAt,
		&order.UpdatedAt,
		&completedAt,
		&cancelledAt,
		&order.BusinessID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if shiftID.Valid {
		sid := int(shiftID.Int64)
		order.ShiftID = &sid
	}

	if completedAt.Valid {
		order.CompletedAt = &completedAt.Time
	}

	if cancelledAt.Valid {
		order.CancelledAt = &cancelledAt.Time
	}

	// Load order items
	items, err := r.GetOrderItems(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	order.Items = items

	return &order, nil
}

// AddOrderItem adds an item to an order
func (r *orderRepository) AddOrderItem(ctx context.Context, orderID int, item *entity.OrderItem) error {
	query := `
		INSERT INTO order_items (order_id, dish_id, quantity, price, notes, created_at, updated_at, business_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	now := time.Now()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = now
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		orderID,
		item.DishID,
		item.Quantity,
		item.Price,
		item.Notes,
		item.CreatedAt,
		item.UpdatedAt,
		item.BusinessID,
	).Scan(&item.ID)

	return err
}

// RemoveOrderItem removes an item from an order
func (r *orderRepository) RemoveOrderItem(ctx context.Context, orderID, itemID int) error {
	query := `DELETE FROM order_items WHERE order_id = $1 AND id = $2`

	result, err := r.db.ExecContext(ctx, query, orderID, itemID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("order item not found")
	}

	return nil
}

// UpdateOrderItem updates an order item
func (r *orderRepository) UpdateOrderItem(ctx context.Context, item *entity.OrderItem) error {
	query := `
		UPDATE order_items
		SET dish_id = $1, quantity = $2, price = $3, notes = $4, updated_at = $5
		WHERE id = $6 AND order_id = $7
	`

	item.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		item.DishID,
		item.Quantity,
		item.Price,
		item.Notes,
		item.UpdatedAt,
		item.ID,
		item.OrderID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("order item not found")
	}

	return nil
}

// GetOrderItems retrieves all items for an order
func (r *orderRepository) GetOrderItems(ctx context.Context, orderID int) ([]*entity.OrderItem, error) {
	query := `
		SELECT id, order_id, dish_id, quantity, price, notes, created_at, updated_at, business_id
		FROM order_items
		WHERE order_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.OrderItem

	for rows.Next() {
		var item entity.OrderItem

		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.DishID,
			&item.Quantity,
			&item.Price,
			&item.Notes,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.BusinessID,
		)

		if err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
