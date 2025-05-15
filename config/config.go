package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Port               string `json:"port"`
	MongoURI           string `json:"mongo_uri"`
	JWTSecret          string `json:"jwt_secret"`
	DefaultUserQuotaMB int64  `json:"default_user_quota_mb"`
}

func LoadConfig(filePath string) (Config, error) {

	var AppConfig Config

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&AppConfig)
	if err != nil {
		log.Fatalf("Error decoding config JSON: %v", err)
	}

	log.Println("Config loaded successfully")

	return AppConfig, nil
}
