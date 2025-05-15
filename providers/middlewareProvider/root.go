package middlewareProvider

import (
	providers "github.com/file_upload/providers"
)

type Middleware struct {
	DBHelper providers.DBHelperProvider
}

func NewMiddleware(dbHelper providers.DBHelperProvider) providers.MiddlewareProvider {
	return &Middleware{
		DBHelper: dbHelper,
	}
}
