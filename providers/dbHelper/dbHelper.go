package dbHelper

import (
	"github.com/file_upload/providers"
	"go.mongodb.org/mongo-driver/mongo"
)

type DBHelper struct {
	UserCollection         *mongo.Collection
	UserSessionsCollection *mongo.Collection
	FileCollection         *mongo.Collection
}

func NewDBHelperProvider(db *mongo.Client) providers.DBHelperProvider {
	return &DBHelper{
		UserCollection:         (*mongo.Collection)(db.Database("WOBOT_AI").Collection("users")),
		FileCollection:         (*mongo.Collection)(db.Database("WOBOT_AI").Collection("files")),
		UserSessionsCollection: (*mongo.Collection)(db.Database("WOBOT_AI").Collection("userSessions")),
	}
}
