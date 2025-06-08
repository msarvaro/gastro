package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"restaurant-management/internal/domain/order"
	"time"

	"github.com/lib/pq"
)

type OrderRepository struct {
	db *DB
}

func NewOrderRepository(db *DB) order.Repository {
	return &OrderRepository{db: db}
}

// GetDishByID retrieves a specific dish by its ID.
func (r *OrderRepository) GetDishByID(ctx context.Context, id int) (*order.Dish, error) {
	dish := &order.Dish{}
	err := r.db.QueryRowContext(ctx, "SELECT id, name, category_id, price, is_available FROM dishes WHERE id = $1", id).
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
func (r *OrderRepository) GetActiveOrdersWithItems(ctx context.Context, businessID int) ([]order.Order, error) {
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
	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		log.Printf("Error in GetActiveOrdersWithItems query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var orders []order.Order
	for rows.Next() {
		var o order.Order
		var itemsJSON []byte
		var completedAt pq.NullTime
		var cancelledAt pq.NullTime

		err := rows.Scan(
			&o.ID, &o.TableID, &o.WaiterID, &o.Status, &o.Comment, &o.TotalAmount,
			&o.CreatedAt, &o.UpdatedAt, &completedAt, &cancelledAt, &itemsJSON,
		)
		if err != nil {
			log.Printf("Error scanning active order row: %v", err)
			return nil, err
		}

		if completedAt.Valid {
			o.CompletedAt = &completedAt.Time
		}
		if cancelledAt.Valid {
			o.CancelledAt = &cancelledAt.Time
		}

		if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
			log.Printf("Error unmarshalling order items for order %d: %v", o.ID, err)
			return nil, err
		}
		orders = append(orders, o)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error after iterating active order rows: %v", err)
		return nil, err
	}
	return orders, nil
}

// GetOrderByID retrieves a specific order by its ID, including its items.
func (r *OrderRepository) GetOrderByID(ctx context.Context, id int, businessID ...int) (*order.Order, error) {
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

	var o order.Order
	var itemsJSON []byte
	var completedAt, cancelledAt pq.NullTime

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
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
func (r *OrderRepository) CreateOrderAndItems(ctx context.Context, o *order.Order, businessID int) (*order.Order, error) {
	tx, err := r.db.Begin()
	if err != nil {
		log.Printf("Error starting transaction for creating order: %v", err)
		return nil, err
	}

	now := time.Now()
	o.CreatedAt = now
	o.UpdatedAt = now

	orderSQL := `INSERT INTO orders (table_id, waiter_id, status, comment, total_amount, created_at, updated_at, business_id)
                 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`
	err = tx.QueryRowContext(ctx, orderSQL, o.TableID, o.WaiterID, o.Status, o.Comment, o.TotalAmount, o.CreatedAt, o.UpdatedAt, businessID).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		tx.Rollback()
		log.Printf("Error inserting order: %v", err)
		return nil, err
	}

	itemSQL := `INSERT INTO order_items (order_id, dish_id, quantity, price, notes, business_id)
                VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	for i := range o.Items {
		item := &o.Items[i]
		err = tx.QueryRowContext(ctx, itemSQL, o.ID, item.DishID, item.Quantity, item.Price, item.Notes, businessID).Scan(&item.ID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction for creating order: %v", err)
		return nil, err
	}
	return o, nil
}

// UpdateOrder updates an existing order's status and relevant timestamps.
// It expects order.Status, and potentially order.CompletedAt or order.CancelledAt to be set.
// order.ID must be valid.
func (r *OrderRepository) UpdateOrder(ctx context.Context, o *order.Order) error {
	query := `
        UPDATE orders 
        SET status = $1, comment = $2, total_amount = $3, 
            updated_at = $4, completed_at = $5, cancelled_at = $6
        WHERE id = $7`

	o.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		o.Status, o.Comment, o.TotalAmount,
		o.UpdatedAt, o.CompletedAt, o.CancelledAt,
		o.ID,
	)
	if err != nil {
		log.Printf("Error updating order ID %d: %v", o.ID, err)
		return err
	}
	return nil
}

func (r *OrderRepository) GetOrderStatus(ctx context.Context, businessID int) (*order.OrderStats, error) {
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

	var stats order.OrderStats
	err := r.db.QueryRowContext(ctx, query, businessID).Scan(
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
func (r *OrderRepository) GetOrderHistoryWithItems(ctx context.Context, businessID int) ([]order.Order, error) {
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

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		log.Printf("Error in GetOrderHistoryWithItems query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var orders []order.Order
	for rows.Next() {
		var o order.Order
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
func (r *OrderRepository) IsLastActiveOrderForTable(ctx context.Context, tableID, currentOrderID int) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
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
func (r *OrderRepository) GetOrdersByStatus(ctx context.Context, status string, businessID int) ([]order.Order, error) {
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
	rows, err := r.db.QueryContext(ctx, query, status, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []order.Order
	for rows.Next() {
		var o order.Order
		var itemsJSON []byte
		var completedAt, cancelledAt sql.NullTime

		err := rows.Scan(
			&o.ID, &o.TableID, &o.WaiterID, &o.Status, &o.Comment, &o.TotalAmount,
			&o.CreatedAt, &o.UpdatedAt, &completedAt, &cancelledAt, &itemsJSON,
		)
		if err != nil {
			return nil, err
		}
		if completedAt.Valid {
			o.CompletedAt = &completedAt.Time
		}
		if cancelledAt.Valid {
			o.CancelledAt = &cancelledAt.Time
		}
		if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}
