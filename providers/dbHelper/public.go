package dbHelper

import (
	"context"
	"fmt"
	"time"

	"github.com/file_upload/models"
	"github.com/file_upload/utils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (dh *DBHelper) GetUserByUsername(username string) (models.User, error) {

	var users models.User

	// filters
	filter := bson.M{"username": username}

	result := dh.UserCollection.FindOne(context.TODO(), filter)

	err := result.Decode(&users)
	if err != nil {
		logrus.Errorf("GetUserByUsername, error decoding the database results : %v ", err)
		return users, err
	}

	return users, nil
}

func (dbHelper *DBHelper) ReadUserSessions(userID string, activeSessions bool) ([]models.UserSession, error) {

	utils.LogInfo("ReadUserSessions", "reading all the User sessions with the specified UserID", fmt.Sprintf("UserID: %s, Active Sessions Switch: %v", userID, activeSessions), nil)

	var userSessionsData []models.UserSession
	var filter primitive.M

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// here bson.M act as a map.
	filter = bson.M{"userId": userID}

	logrus.Infof("Mongo Filter 1: %v", filter)

	if activeSessions {
		filter["endTime"] = bson.M{"$gt": time.Now().Unix()}
	}

	logrus.Infof("Mongo Filter 2: %v", filter)

	cursor, err := dbHelper.UserSessionsCollection.Find(ctx, &filter)
	if err != nil {
		utils.LogError("ReadUserSessions", "error reading user session data from the database", fmt.Sprintf("UserID: %s, Count of User Sessions: %d, Active Sessions Switch: %v", userID, len(userSessionsData), activeSessions), err)
		return userSessionsData, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &userSessionsData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogError("ReadUserSessions", "no documents in the collection", fmt.Sprintf("UserID: %s, Count of User Sessions: %d, Active Sessions Switch: %v", userID, len(userSessionsData), activeSessions), err)
			return userSessionsData, err
		}
		utils.LogError("ReadUserSessions", "error decoding user session data from the database", fmt.Sprintf("UserID: %s, Count of User Sessions: %d, Active Sessions Switch: %v", userID, len(userSessionsData), activeSessions), err)
		return userSessionsData, err
	}
	utils.LogInfo("ReadUserSessions", "successfully read user sessions from the database", fmt.Sprintf("UserID: %s, Count of User Sessions: %d, Active Sessions Switch: %v", userID, len(userSessionsData), activeSessions), nil)

	return userSessionsData, nil
}

func (dbHelper *DBHelper) ReadUserSessionBySessionID(sessionID string) (models.UserSession, error) {

	utils.LogInfo("ReadUserSessionBySessionID", "reading the User session with the specified ID", fmt.Sprintf("SessionID: %s", sessionID), nil)

	var userSession models.UserSession

	filter := bson.M{"id": sessionID}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := dbHelper.UserSessionsCollection.FindOne(ctx, &filter).Decode(&userSession)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("EndUserSession", "no sessions to be fetched", fmt.Sprintf("SessionID: %s", sessionID), err)
			return userSession, nil
		}
		utils.LogError("ReadUserSessionBySessionID", "error decoding user session data from the database", fmt.Sprintf("SessionID: %s", sessionID), err)
		return userSession, err
	}
	utils.LogInfo("ReadUserSessionBySessionID", "user session fetched successfully", fmt.Sprintf("SessionID: %s", sessionID), nil)

	return userSession, nil
}

func (dbHelper *DBHelper) EndUserSession(sessionID string) error {

	utils.LogInfo("EndUserSession", "ending the User session with the specified ID", fmt.Sprintf("SessionID: %s", sessionID), nil)

	userSessionData, err := dbHelper.ReadUserSessionBySessionID(sessionID)
	if err != nil {
		utils.LogError("EndUserSession", "error fetching user session data", fmt.Sprintf("SessionID: %s", sessionID), err)
		return err
	}

	filter := bson.M{"id": sessionID}
	update := bson.M{"$set": bson.M{"endTime": time.Now().Unix()}}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := dbHelper.UserSessionsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("EndUserSession", "no sessions to be closed", fmt.Sprintf("SessionID: %s", sessionID), err)
			return nil
		}
		utils.LogError("EndUserSession", "error ending the user session in the database", fmt.Sprintf("SessionID: %s", sessionID), err)
		return err
	}
	utils.LogInfo("EndUserSession", fmt.Sprintf("successfully ended the user session in the database. Matched Count: %d, Modified Count: %d, Upserted Count: %d", result.MatchedCount, result.ModifiedCount, result.UpsertedCount), "", userSessionData)

	return nil
}

func (dbHelper *DBHelper) CreateUserSession(userID string) (models.UserSession, error) {

	utils.LogInfo("CreateUserSession", "create a new User session based on the data provided", fmt.Sprintf("UserID: %s", userID), nil)

	var newSession models.UserSession

	userSessions, err := dbHelper.ReadUserSessions(userID, true)
	if err != nil {
		utils.LogError("CreateUserSession", "error fetching user's sessions data from the database", userID, err)
		return newSession, err
	}

	// end all the sessions.
	if len(userSessions) > 0 {

		fmt.Println("Hello I am here")
		for _, userSession := range userSessions {

			err = dbHelper.EndUserSession(userSession.ID)
			if err != nil {
				utils.LogError("CreateUserSession", "error closing user's previous sessions in the database", fmt.Sprintf("User ID: %s, Old User Session ID: %s", userID, userSession.ID), err)
				return newSession, err
			}
		}

		utils.LogInfo("CreateUserSession", "successfully end all the session", "", "")

	}

	newSession.ID = uuid.New().String()
	newSession.UserID = userID
	newSession.StartTime = time.Now().Unix()
	newSession.EndTime = time.Now().Add(1 * time.Hour).Unix()
	newSession.Token = uuid.New().String()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := dbHelper.UserSessionsCollection.InsertOne(ctx, newSession)
	if err != nil {
		utils.LogError("CreateUserSession", "error inserting user's new session into the database", fmt.Sprintf("UserID: %s", userID), err)
		return newSession, err
	}
	utils.LogInfo("CreateUserSession", "successfully created a new user session", fmt.Sprintf("New Session ID: %s", newSession.ID), result.InsertedID)

	return newSession, nil
}

func (dh *DBHelper) CreateUser(user models.User) error {

	filter := bson.M{"username": user.Username}

	var existingUser models.User
	err := dh.UserCollection.FindOne(context.TODO(), filter).Decode(&existingUser)
	if err == nil {
		logrus.Warnf("CreateUser: user with username '%s' already exists", user.Username)
		return fmt.Errorf("user with username '%s' already exists", user.Username)
	}
	if err != mongo.ErrNoDocuments {
		logrus.Errorf("CreateUser: error checking for existing user: %v", err)
		return err
	}
	_, err = dh.UserCollection.InsertOne(context.TODO(), user)
	if err != nil {
		logrus.Errorf("CreateUser, error inserting user data in mongo database users collection : %v ", err)
		return err
	}
	return nil
}

func (dbHelper *DBHelper) IsUserSessionTokenActive(tokenString string) (bool, error) {

	utils.LogInfo("IsUserSessionTokenActive", "checking if user token is valid in the database", fmt.Sprintf("Token: %s", tokenString), nil)

	sessionData, err := dbHelper.ReadUserSessionBySessionToken(tokenString)
	if err != nil {
		utils.LogError("IsUserSessionTokenActive", "error reading user session data from the database, session is assumed to be inactive", fmt.Sprintf("Token: %s", tokenString), err)
		return false, nil
	}

	if sessionData.EndTime < time.Now().Unix() {
		utils.LogInfo("IsUserSessionTokenActive", "token has been expired", fmt.Sprintf("Token: %s", tokenString), nil)
		return false, nil
	}

	utils.LogInfo("IsUserSessionTokenActive", "token data found valid for an active session", fmt.Sprintf("Token: %s", tokenString), nil)

	return true, nil
}

func (dbHelper *DBHelper) UpdateUserSession(sessionID string) error {

	utils.LogInfo("UpdateUserSession", "updating the User session with the specified session ID", fmt.Sprintf("SessionID: %s", sessionID), nil)

	userSessionData, err := dbHelper.ReadUserSessionBySessionID(sessionID)
	if err != nil {
		utils.LogError("UpdateUserSession", "error fetching user session data", fmt.Sprintf("SessionID: %s", sessionID), err)
		return err
	}

	userSessionData.EndTime = time.Now().Add(1 * time.Hour).Unix()

	filter := bson.M{"id": sessionID}
	update := bson.M{"$set": userSessionData}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := dbHelper.UserSessionsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		utils.LogError("UpdateUserSession", "error updating user session in the database", fmt.Sprintf("SessionID: %s", sessionID), err)
		return err
	}
	utils.LogInfo("UpdateUserSession", fmt.Sprintf("successfully updated user session in the database. Matched Count: %d, Modified Count: %d, Upserted Count: %d", result.MatchedCount, result.ModifiedCount, result.UpsertedCount), fmt.Sprintf("SessionID: %s", sessionID), userSessionData)

	return nil
}

func (dbHelper *DBHelper) IsUserSessionActive(sessionID string) (bool, error) {

	utils.LogInfo("IsUserSessionActive", "checking if user is marked online in the database", fmt.Sprintf("SessionID: %s", sessionID), nil)

	sessionData, err := dbHelper.ReadUserSessionBySessionID(sessionID)
	if err != nil {
		utils.LogError("IsUserSessionActive", "error reading user session data from the database, session is assumed to be inactive", fmt.Sprintf("SessionID: %s", sessionID), err)
		return false, nil
	}

	if sessionData.EndTime <= time.Now().Unix() {
		utils.LogInfo("IsUserSessionActive", "session has been terminated", fmt.Sprintf("SessionID: %s", sessionID), nil)
		return false, nil
	}

	utils.LogInfo("IsUserSessionActive", "active session data found for the session ID, session is active", fmt.Sprintf("SessionID: %s", sessionID), nil)

	return true, nil
}

func (dbHelper *DBHelper) ReadUserSessionBySessionToken(tokenString string) (models.UserSession, error) {

	utils.LogInfo("ReadUserSessionBySessionToken", "reading the User session with the specified token", fmt.Sprintf("Token: %s", tokenString), nil)

	var userSession models.UserSession

	filter := bson.M{"token": tokenString}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := dbHelper.UserSessionsCollection.FindOne(ctx, &filter).Decode(&userSession)
	if err != nil {
		utils.LogError("ReadUserSessionBySessionToken", "error decoding user session data from the database", fmt.Sprintf("Token: %s", tokenString), err)
		return userSession, err
	}
	utils.LogInfo("ReadUserSessionBySessionToken", "user session fetched successfully", fmt.Sprintf("Token: %s", tokenString), nil)

	return userSession, nil
}

func (dh *DBHelper) GetUserByID(userID string) (models.User, error) {
	utils.LogInfo("GetUserByID", "fetching user by ID", fmt.Sprintf("UserID: %s", userID), nil)

	var user models.User
	filter := bson.M{"id": userID}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := dh.UserCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		utils.LogError("GetUserByID", "error decoding the user from the database", fmt.Sprintf("UserID: %s", userID), err)
		return user, err
	}

	utils.LogInfo("GetUserByID", "user fetched successfully", fmt.Sprintf("UserID: %s", userID), nil)
	return user, nil
}

func (dh *DBHelper) UpdateStorageData(userID string, storage int64) error {
	utils.LogInfo("UpdateStorageData", "updating used storage", fmt.Sprintf("UserID: %s, NewStorage: %d", userID, storage), nil)

	filter := bson.M{"id": userID}
	update := bson.M{"$set": bson.M{"used_storage": storage}}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := dh.UserCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		utils.LogError("UpdateStorageData", "error updating user storage in the database", fmt.Sprintf("UserID: %s", userID), err)
		return err
	}

	utils.LogInfo("UpdateStorageData", fmt.Sprintf("storage updated. Matched: %d, Modified: %d", result.MatchedCount, result.ModifiedCount), fmt.Sprintf("UserID: %s", userID), nil)
	return nil
}

func (dh *DBHelper) InsertFileMetadata(file models.File) error {
	utils.LogInfo("InsertFileMetadata", "inserting file metadata", fmt.Sprintf("UserID: %s, FileName: %s", file.UserID, file.Filename), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := dh.FileCollection.InsertOne(ctx, file)
	if err != nil {
		utils.LogError("InsertFileMetadata", "error inserting file metadata", fmt.Sprintf("UserID: %s, FileName: %s", file.UserID, file.Filename), err)
	}
	return err
}

func (dh *DBHelper) GetFileByHash(userID, hash string) (*models.File, error) {
	utils.LogInfo("GetFileByHash", "searching for file by hash", fmt.Sprintf("UserID: %s, Hash: %s", userID, hash), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var file models.File
	err := dh.FileCollection.FindOne(ctx, bson.M{"user_id": userID, "hash": hash}).Decode(&file)
	if err != nil {
		utils.LogError("GetFileByHash", "file not found or error decoding", fmt.Sprintf("UserID: %s, Hash: %s", userID, hash), err)
		return nil, err
	}

	utils.LogInfo("GetFileByHash", "file retrieved successfully", fmt.Sprintf("UserID: %s, FileName: %s", userID, file.Filename), nil)
	return &file, nil
}

func (dh *DBHelper) GetFilesByUser(userID string) ([]models.File, error) {
	utils.LogInfo("GetFilesByUser", "fetching all files for user", fmt.Sprintf("UserID: %s", userID), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := dh.FileCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		utils.LogError("GetFilesByUser", "error fetching files from database", fmt.Sprintf("UserID: %s", userID), err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var files []models.File
	if err = cursor.All(ctx, &files); err != nil {
		utils.LogError("GetFilesByUser", "error decoding file cursor", fmt.Sprintf("UserID: %s", userID), err)
		return nil, err
	}

	utils.LogInfo("GetFilesByUser", fmt.Sprintf("retrieved %d files", len(files)), fmt.Sprintf("UserID: %s", userID), nil)
	return files, nil
}
