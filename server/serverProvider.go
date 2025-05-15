package server

import (
	"context"
	"net/http"
	"time"

	"github.com/file_upload/config"
	"github.com/file_upload/models"
	"github.com/file_upload/providers"
	"github.com/file_upload/providers/dbHelper"
	"github.com/file_upload/providers/dbProvider"
	middlewareprovider "github.com/file_upload/providers/middlewareProvider"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	MongoClient        *mongo.Database
	DBHelper           providers.DBHelperProvider
	httpServer         *http.Server
	MiddlewareProvider providers.MiddlewareProvider
	Config             *config.Config
}

func SrvInit(config *config.Config) *Server {
	mongoClient := dbProvider.ConnectDB(config.MongoURI, models.ConnectDBMaxAttempts)
	if mongoClient == nil {
		logrus.Error("Server Init: Failed to establish client connection to database, Error establishing database connection")
	}

	dbHelper := dbHelper.NewDBHelperProvider(mongoClient.Client())

	middleWare := middlewareprovider.NewMiddleware(dbHelper)

	return &Server{
		DBHelper:           dbHelper,
		MiddlewareProvider: middleWare,
		Config:             config,
	}

}

func (srv *Server) Start() {
	addr := ":" + srv.Config.Port
	httpServ := &http.Server{
		Addr:    addr,
		Handler: srv.InjectRoutes(),
	}

	srv.httpServer = httpServ
	logrus.Info("Server running at PORT ", addr)

	if err := httpServ.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Fatalf("Start %v", err)
		return
	}
}

func (srv *Server) Stop() {
	logrus.Info("closing Server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logrus.Info("closing server...")
	_ = srv.httpServer.Shutdown(ctx)
	logrus.Info("Done")
}
