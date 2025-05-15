package dbProvider

import (
	"context"
	"time"

	"github.com/file_upload/providers"
	"github.com/file_upload/utils"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoClientProvider struct {
	mongoClient *mongo.Client
	ctx         context.Context
}

func ConnectDB(address string, maxAttempts int) providers.MongoClientProvider {

	var (
		mongoClient *mongo.Client
		err         error
	)

	ctx := context.Background()

	for i := 0; i < maxAttempts; i++ {

		mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(address))
		if err != nil {
			logrus.Errorf("dbProvider.go: ConnectDB: Failed to establish a MongoDB client: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// TODO: If code reach here then it means connection is establish.
		// But I will look into this.
		logrus.Info("dbProvider.go: ConnectDB: MongoDB client connected to the database successfully.")
		break
	}

	if err == nil {
		return &mongoClientProvider{
			mongoClient: mongoClient,
			ctx:         ctx,
		}
	} else {
		return nil
	}
}

func (mongoClientProviderObject *mongoClientProvider) Ping() error {
	return mongoClientProviderObject.mongoClient.Ping(mongoClientProviderObject.ctx, nil)
}

func (mongoClientProviderObject *mongoClientProvider) Client() *mongo.Client {
	return mongoClientProviderObject.mongoClient
}

func (mongoClientProviderObject *mongoClientProvider) DisconnectDB() error {

	// TODO: I need to make function for fetch the context from the mongoClient or provider where we store.
	ctx := mongoClientProviderObject.Context()
	err := mongoClientProviderObject.mongoClient.Disconnect(ctx)
	if err != nil {
		utils.LogError("dbProvider.go", "DisconnectDB: MongoDB client failed to disconnect from the database.", "", err)
		return err
	}

	return nil
}

func (mongoClientProviderObject *mongoClientProvider) Context() context.Context {
	return mongoClientProviderObject.ctx
}
