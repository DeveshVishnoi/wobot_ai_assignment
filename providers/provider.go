package providers

import (
	"context"

	"github.com/file_upload/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type DBHelperProvider interface {
	GetUserByUsername(string) (models.User, error)
	CreateUserSession(string) (models.UserSession, error)
	CreateUser(models.User) error
	GetUserByID(userID string) (models.User, error)
	UpdateStorageData(string, int64) error

	IsUserSessionActive(sessionID string) (bool, error)
	UpdateUserSession(sessionID string) error
	IsUserSessionTokenActive(tokenString string) (bool, error)
	ReadUserSessionBySessionToken(tokenString string) (models.UserSession, error)

	InsertFileMetadata(models.File) error
	GetFileByHash(string, string) (*models.File, error)
	GetFilesByUser(string) ([]models.File, error)
}

type MiddlewareProvider interface {
	AuthMiddleware() gin.HandlerFunc
	UserFromContext(ctx context.Context) *models.UserContext
}

type MongoClientProvider interface {

	// ping check the connection with the db.
	Ping() error

	// disconnect the conection with db.
	DisconnectDB() error

	// contect find. (use in the pinging the connection)
	Context() context.Context

	// gives mongio client pointer.
	Client() *mongo.Client
}
