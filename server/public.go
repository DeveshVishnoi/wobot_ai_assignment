package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/file_upload/providers/authProvider"
	"github.com/file_upload/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/file_upload/models"
	"github.com/gin-gonic/gin"
)

func (srv *Server) login(c *gin.Context) {

	var usernameAndPassword models.UsernameAndPassword

	err := json.NewDecoder(c.Request.Body).Decode(&usernameAndPassword)
	if err != nil {
		utils.LogError("login", "error decoding request body", "", err)
		utils.RespondClientErr(c, err, http.StatusBadRequest, "error decoding request body")
		return
	}

	userDetail, err := srv.DBHelper.GetUserByUsername(usernameAndPassword.Username)
	if err != nil {
		utils.LogError("login", "error invalid username", usernameAndPassword, err)
		utils.RespondGenericServerErr(c, err, "error invalid username")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userDetail.Password), []byte(usernameAndPassword.Password)); err != nil {
		utils.LogError("login", "invalid password", usernameAndPassword, err)
		utils.RespondClientErr(c, err, http.StatusBadRequest, "invalid password")
		return
	}

	session, err := srv.DBHelper.CreateUserSession(userDetail.ID)
	if err != nil {
		utils.LogError("login", "error creating user session", usernameAndPassword, err)
		utils.RespondGenericServerErr(c, err, "error creating user session")
		return
	}

	token, err := authProvider.GenerateJWT(userDetail, session.Token)
	if err != nil {
		utils.LogError("login", "error creating user's auth token", usernameAndPassword, err)
		utils.RespondGenericServerErr(c, err, "error creating user's auth token")
		return
	}

	utils.EncodeJSONBody(c, http.StatusOK, map[string]interface{}{
		"token":  token,
		"userID": userDetail.ID,
	})
}

func (srv *Server) createNewUser(c *gin.Context) {

	var user models.User

	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		utils.LogError("createNewUser", "error decoding request body", "", err)
		utils.RespondClientErr(c, err, http.StatusBadRequest, "error decoding request body")
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hash)
	user.ID = uuid.NewString()
	user.CreatedAt = time.Now().Unix()
	user.Quota = srv.Config.DefaultUserQuotaMB * 1024 * 1024

	err = srv.DBHelper.CreateUser(user)
	if err != nil {
		utils.LogError("createNewUser", "error inserting user in the server database", user, err)
		utils.RespondGenericServerErr(c, err, "error inserting user in the server database")
		return
	}

	utils.EncodeJSONBody(c, http.StatusOK, map[string]interface{}{
		"message": "successfully added user",
	})
}

func (srv *Server) remainingStorage(c *gin.Context) {
	user := srv.MiddlewareProvider.UserFromContext(c.Request.Context())

	utils.EncodeJSONBody(c, http.StatusOK, map[string]interface{}{
		"total":     user.Quota,
		"used":      user.UsedStorage,
		"remaining": user.Quota - user.UsedStorage,
	})
}

func (srv *Server) uploadFile(c *gin.Context) {

	userContext := srv.MiddlewareProvider.UserFromContext(c.Request.Context())

	dirPath := fmt.Sprintf("%s/%s", models.DefaultDirectory, userContext.Username)
	err := utils.CreateDirIfNotExist(dirPath)
	if err != nil {
		log.Fatalf("Error ensuring directory exists: %v", err)
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.LogError("uploadFile", "error getting file from form", "", err)
		utils.RespondClientErr(c, err, http.StatusBadRequest, "file not found in form data")
		return
	}
	defer file.Close()

	fmt.Println("file -", file, header.Size, header.Filename)

	// TODO: checking if any files are uploading again.
	// Reason: the memroy will increase but file will not add in the directory, or we will do one thing, change the name of the file.
	var fileBuffer bytes.Buffer
	tee := io.TeeReader(file, &fileBuffer)

	hash := sha256.New()
	if _, err := io.Copy(hash, tee); err != nil {
		utils.RespondGenericServerErr(c, err, "failed to compute file hash")
		return
	}
	fileHash := fmt.Sprintf("%x", hash.Sum(nil))

	// check anmy file are present with same hash or not.
	existingFile, err := srv.DBHelper.GetFileByHash(userContext.ID, fileHash)
	if err == nil && existingFile != nil {
		utils.RespondClientErr(c, fmt.Errorf("duplicate file"), http.StatusConflict, "file already uploaded")
		return
	}

	fileReader := bytes.NewReader(fileBuffer.Bytes())

	if header.Size+userContext.UsedStorage > userContext.Quota {
		utils.RespondClientErr(c, fmt.Errorf("alert, User don't have storage to store the file :%v, size: %v, you want ", header.Filename, header.Size), http.StatusBadRequest, "insufficient Storage")
		return
	}

	uploadPath := fmt.Sprintf("%s/%s", dirPath, header.Filename)
	outFile, err := os.Create(uploadPath)
	if err != nil {
		utils.LogError("uploadFile", "error saving file", "", err)
		utils.RespondGenericServerErr(c, err, "unable to save uploaded file")
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, fileReader)
	if err != nil {
		utils.LogError("uploadFile", "error copying file data", "", err)
		utils.RespondGenericServerErr(c, err, "error while saving file")
		return
	}

	err = srv.DBHelper.UpdateStorageData(userContext.ID, header.Size+userContext.UsedStorage)
	if err != nil {
		utils.RespondGenericServerErr(c, err, "error while update the storage the file")
		return
	}

	newFile := models.File{
		ID:         uuid.NewString(),
		UserID:     userContext.ID,
		Filename:   header.Filename,
		Size:       header.Size,
		Path:       uploadPath,
		Hash:       fileHash,
		UploadedAt: time.Now().Unix(),
	}

	if err := srv.DBHelper.InsertFileMetadata(newFile); err != nil {
		utils.RespondGenericServerErr(c, err, "failed to save file metadata")
		return
	}

	utils.EncodeJSONBody(c, http.StatusOK, map[string]interface{}{
		"message":  "file uploaded successfully",
		"filename": header.Filename,
		"userID":   userContext.ID,
	})
}

func (srv *Server) getUserFiles(c *gin.Context) {
	userContext := srv.MiddlewareProvider.UserFromContext(c.Request.Context())

	files, err := srv.DBHelper.GetFilesByUser(userContext.ID)
	if err != nil {
		utils.LogError("getUserFiles", "fetching user files", "", err)
		utils.RespondGenericServerErr(c, err, "could not retrieve user files")
		return
	}

	utils.EncodeJSONBody(c, http.StatusOK, map[string]interface{}{
		"user_id": userContext.ID,
		"files":   files,
	})
}
