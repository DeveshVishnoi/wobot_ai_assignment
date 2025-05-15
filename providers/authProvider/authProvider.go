package authProvider

import (
	"strconv"
	"time"

	"github.com/file_upload/models"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

func GenerateJWT(user models.User, sessionToken string) (tokenString string, err error) {

	claims := &jwt.MapClaims{
		"iss": user.ID,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"data": map[string]string{
			"id":        user.ID,
			"username":  user.Username,
			"expiresAt": strconv.Itoa(int(time.Now().Add(1 * time.Hour).Unix())),
			"token":     sessionToken,
			"issuer":    user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err = token.SignedString(models.JwtSigningSecretKey)
	if err != nil {
		logrus.Errorf("GenerateJWT: error signing the token: %v ", err)
		// utils.LogError(err)
		return tokenString, err
	}

	return tokenString, nil
}
