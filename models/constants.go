package models

const (
	UserContextKey string = "userContext"

	ConnectDBMaxAttempts = 3

	// Middleware
	MiddlewareBearerScheme = "bearer"
	MiddlewareSpace        = " "

	// server Error Message.
	ServerErrorMsg   = "Internal Server Error occurred. Please contact your administrator."
	DefaultDirectory = "storage"
)

var JwtSigningSecretKey = []byte("supersecretkey")
