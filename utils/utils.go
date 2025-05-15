package utils

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/file_upload/models"
	"github.com/gin-gonic/gin"
)

func GetModuleDirectoryPath() (string, error) {

	executablePath, err := os.Executable()
	if err != nil {
		LogFatal("GetModuleDirectoryPath", "error getting the path of the executable", "", err)
		return "", err
	}
	return filepath.Dir(executablePath), nil
}

func RespondClientErr(c *gin.Context, err error, statusCode int, messageToUser string, additionalInfoForDevs ...string) {

	additionalInfoJoined := strings.Join(additionalInfoForDevs, "\n")

	data := models.ClientError{
		MessageToUser: messageToUser,
		DeveloperInfo: additionalInfoJoined,
		Err:           err.Error(),
		StatusCode:    statusCode,
		IsClientError: true,
	}

	c.JSON(statusCode, data)
}

func RespondGenericServerErr(c *gin.Context, err error, additionalInfoForDevs ...string) {

	additionalInfoJoined := strings.Join(additionalInfoForDevs, "\n")

	data := models.ClientError{
		MessageToUser: models.ServerErrorMsg,
		DeveloperInfo: additionalInfoJoined,
		Err:           err.Error(),
		StatusCode:    http.StatusInternalServerError,
		IsClientError: false,
	}

	c.JSON(http.StatusInternalServerError, data)
}

func EncodeJSONBody(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

func CreateDirIfNotExist(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			log.Printf("Failed to create directory: %s, error: %v\n", path, err)
			return err
		}
		log.Printf("Directory created: %s\n", path)
	} else if err != nil {
		log.Printf("Error checking directory: %v\n", err)
		return err
	} else {
		log.Printf("Directory already exists: %s\n", path)
	}
	return nil
}

func init() {
	err := CreateDirIfNotExist(models.DefaultDirectory)
	if err != nil {
		log.Fatalf("Error ensuring directory exists: %v", err)
	}
}
