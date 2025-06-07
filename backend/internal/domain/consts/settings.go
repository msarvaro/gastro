package consts

// Default settings
const (
	DefaultPageSize        = 20
	MaxPageSize            = 100
	DefaultCurrency        = "USD"
	DefaultLanguage        = "en"
	DefaultTimezone        = "UTC"
	TokenExpiryHours       = 24
	RefreshTokenExpiryDays = 30
)

// Time formats
const (
	DateFormat     = "2006-01-02"
	TimeFormat     = "15:04:05"
	DateTimeFormat = "2006-01-02 15:04:05"
)

// Business hours format
const (
	BusinessHourFormat = "15:04"
)

// Performance metrics
const (
	ExcellentPerformance = 90.0
	GoodPerformance      = 75.0
	AveragePerformance   = 60.0
	PoorPerformance      = 40.0
)

// Inventory thresholds
const (
	CriticalStockPercentage = 10.0 // Below 10% of minimum stock
	LowStockPercentage      = 25.0 // Below 25% of minimum stock
	ExpiryWarningDays       = 3    // Warn 3 days before expiry
)

// Order time limits (in minutes)
const (
	OrderAcknowledgeTime   = 5  // Time to acknowledge order
	AveragePreparationTime = 20 // Average dish preparation time
	MaxWaitTime            = 45 // Maximum acceptable wait time
)
