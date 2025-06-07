package consts

// User roles
const (
	RoleAdmin   = "admin"   // System administrator
	RoleManager = "manager" // Business manager
	RoleWaiter  = "waiter"  // Waiter/Service staff
	RoleKitchen = "kitchen" // Kitchen staff/Chef
	RoleCashier = "cashier" // Cashier
)

// Role permissions
var RolePermissions = map[string][]string{
	RoleAdmin: {
		"manage_businesses",
		"manage_all_users",
		"view_all_reports",
		"manage_system_settings",
	},
	RoleManager: {
		"manage_business_users",
		"manage_menu",
		"manage_inventory",
		"view_reports",
		"manage_suppliers",
		"manage_shifts",
	},
	RoleWaiter: {
		"take_orders",
		"manage_tables",
		"view_menu",
		"handle_requests",
		"view_own_performance",
	},
	RoleKitchen: {
		"view_orders",
		"update_order_items",
		"manage_dish_availability",
		"view_inventory",
	},
	RoleCashier: {
		"process_payments",
		"view_orders",
		"generate_bills",
		"handle_refunds",
	},
}
