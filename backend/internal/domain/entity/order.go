package entity

import "time"

// Order represents a customer order
type Order struct {
	ID          int
	TableID     int
	WaiterID    int
	ShiftID     *int
	Status      string // "new", "accepted", "preparing", "ready", "served", "completed", "cancelled"
	TotalAmount int
	Comment     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
	CancelledAt *time.Time
	BusinessID  int
	Items       []*OrderItem
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID         int
	OrderID    int
	DishID     int
	Quantity   int
	Price      float64
	Notes      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	BusinessID int
	Dish       *Dish             // Reference to the dish
	Options    []*MenuItemOption // Selected options for this item
}

// OrderItemOption represents an option selected for an order item
type OrderItemOption struct {
	ID          int
	OrderItemID int
	OptionID    int
	Option      *MenuItemOption // Reference to the option
}

// IsActive checks if the order is still in progress
func (o *Order) IsActive() bool {
	activeStatuses := map[string]bool{
		"new":       true,
		"accepted":  true,
		"preparing": true,
		"ready":     true,
		"served":    true,
	}

	return activeStatuses[o.Status]
}

// IsCompleted checks if the order is completed
func (o *Order) IsCompleted() bool {
	return o.Status == "completed"
}

// IsCancelled checks if the order is cancelled
func (o *Order) IsCancelled() bool {
	return o.Status == "cancelled"
}

// CalculateTotal recalculates the total amount for the order
func (o *Order) CalculateTotal() int {
	var total float64

	for _, item := range o.Items {
		itemTotal := item.Price * float64(item.Quantity)
		total += itemTotal
	}

	return int(total)
}

// UpdateStatus changes the order status and updates related timestamps
func (o *Order) UpdateStatus(status string) {
	now := time.Now()
	o.Status = status
	o.UpdatedAt = now

	// Update specific timestamps based on status
	switch status {
	case "completed":
		o.CompletedAt = &now
	case "cancelled":
		o.CancelledAt = &now
	}
}

// Business methods
func (o *Order) CanBeModified() bool {
	return o.Status == "new" || o.Status == "accepted"
}
