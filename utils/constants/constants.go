package constant

var (
	// HTTP Status Codes
	SUCCESS              = 200
	CREATED              = 201
	BADREQUEST           = 400
	UNAUTHORIZED         = 401
	FORBIDDEN            = 403
	NOTFOUND             = 404
	CONFLICT             = 409
	INTERNALSERVERERROR  = 500

	// Generic Status Strings
	PENDING         = "PENDING"
	INVALID_REQUEST = "INVALID_REQUEST"
	UN_AUTHORIZED   = "UNAUTHORIZED"
	INTERNAL_ERROR  = "INTERNAL_ERROR"
	PAID            = "PAID"
	FAILED          = "FAILED"

	// Optional: Payment or process states
	PROCESSING = "PROCESSING"
	CANCELLED  = "CANCELLED"
	PLACED     = "PLACED"
	SHIPPED    = "SHIPPED"
	DELIVERED  = "DELIVERED"
)
