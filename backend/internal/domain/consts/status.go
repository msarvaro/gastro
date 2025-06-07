package consts

// Order statuses
const (
	OrderStatusPending   = "pending"   // Order created but not confirmed
	OrderStatusConfirmed = "confirmed" // Order confirmed by kitchen
	OrderStatusPreparing = "preparing" // Order being prepared
	OrderStatusReady     = "ready"     // Order ready for delivery
	OrderStatusDelivered = "delivered" // Order delivered to table
	OrderStatusPaid      = "paid"      // Order paid
	OrderStatusCanceled  = "canceled"  // Order canceled
)

// Order item statuses
const (
	OrderItemStatusPending   = "pending"   // Item waiting to be prepared
	OrderItemStatusPreparing = "preparing" // Item being prepared
	OrderItemStatusReady     = "ready"     // Item ready for delivery
	OrderItemStatusDelivered = "delivered" // Item delivered to table
	OrderItemStatusCanceled  = "canceled"  // Item canceled
)

// Table statuses
const (
	TableStatusAvailable   = "available"   // Table is free
	TableStatusOccupied    = "occupied"    // Table has customers
	TableStatusReserved    = "reserved"    // Table is reserved
	TableStatusMaintenance = "maintenance" // Table under maintenance
)

// User statuses
const (
	UserStatusActive   = "active"   // User is active
	UserStatusInactive = "inactive" // User is inactive
)

// Purchase order statuses
const (
	PurchaseOrderStatusDraft     = "draft"     // Order being created
	PurchaseOrderStatusSent      = "sent"      // Order sent to supplier
	PurchaseOrderStatusConfirmed = "confirmed" // Order confirmed by supplier
	PurchaseOrderStatusDelivered = "delivered" // Order delivered
	PurchaseOrderStatusCanceled  = "canceled"  // Order canceled
)

// Service request statuses
const (
	RequestStatusPending      = "pending"      // Request created
	RequestStatusAcknowledged = "acknowledged" // Request seen by waiter
	RequestStatusCompleted    = "completed"    // Request fulfilled
)

// Request types
const (
	RequestTypeCallWaiter = "call_waiter" // Customer needs waiter
	RequestTypeBill       = "bill"        // Customer wants bill
	RequestTypeWater      = "water"       // Water request
	RequestTypeNapkins    = "napkins"     // Napkins request
	RequestTypeHelp       = "help"        // General help
	RequestTypeComplaint  = "complaint"   // Customer complaint
)

// Request priorities
const (
	RequestPriorityLow    = "low"
	RequestPriorityMedium = "medium"
	RequestPriorityHigh   = "high"
	RequestPriorityUrgent = "urgent"
)

// Stock movement types
const (
	StockMovementIn         = "in"         // Stock added
	StockMovementOut        = "out"        // Stock used
	StockMovementAdjustment = "adjustment" // Manual adjustment
	StockMovementWaste      = "waste"      // Wastage/Spoilage
	StockMovementTransfer   = "transfer"   // Transfer between locations
)

// Payment terms
const (
	PaymentTermsCOD     = "COD"     // Cash on delivery
	PaymentTermsNet15   = "Net 15"  // Payment within 15 days
	PaymentTermsNet30   = "Net 30"  // Payment within 30 days
	PaymentTermsPrepaid = "Prepaid" // Payment in advance
)

// Reservation statuses
const (
	ReservationStatusPending   = "pending"   // Reservation is pending confirmation
	ReservationStatusConfirmed = "confirmed" // Reservation is confirmed
	ReservationStatusCancelled = "cancelled" // Reservation has been cancelled
	ReservationStatusCompleted = "completed" // Reservation period has passed and guest attended (or similar meaning)
	ReservationStatusNoShow    = "no_show"   // Guest did not show up for the reservation
)
