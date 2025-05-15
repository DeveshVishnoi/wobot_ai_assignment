package middlewareProvider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/file_upload/models"
	"github.com/file_upload/providers"
	"github.com/file_upload/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

// AuthenticationMiddleware checks each request token increase the expiration time
func (authMiddleware Middleware) AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		var token string
		var userContextData models.UserContext

		tokenParts := strings.Split(c.Request.Header.Get("Authorization"), models.MiddlewareSpace)
		if len(tokenParts) != 2 {
			utils.RespondClientErr(c, errors.New("token not Bearer"), http.StatusUnauthorized, "invalid token", "invalid token")
			utils.LogDebug("AuthenticationMiddleware", "token length check", "Check token  ", tokenParts)
			c.Abort()
			return
		}

		if !strings.EqualFold(tokenParts[0], models.MiddlewareBearerScheme) {
			utils.RespondClientErr(c, errors.New("token not Bearer"), http.StatusUnauthorized, "invalid token", "invalid token")
			c.Abort()
			return
		}
		token = tokenParts[1]

		claims, err := GetClaimsFromToken(token)
		if err != nil {
			utils.LogError("AuthenticationMiddleware", "fetch claims from auth token", "", err)
			utils.RespondClientErr(c, err, http.StatusUnauthorized, "invalid token", "invalid token")
			c.Abort()
			return
		}

		sessionID, isClaimsVerified, err := getUserDataFromClaims(authMiddleware.DBHelper, claims)
		if err != nil {
			utils.LogError("AuthenticationMiddleware", "fetch user data from claim", "", err)
			utils.RespondClientErr(c, err, http.StatusUnauthorized, "invalid token", "invalid token")
			c.Abort()
			return
		}

		// rejecting request because token and session are expired.
		if !isClaimsVerified {
			utils.LogError("AuthenticationMiddleware", "verifying claim of the token", "Claim verification", err)
			utils.RespondClientErr(c, errors.New("invalid token"), http.StatusUnauthorized, "invalid token", "invalid token")
			c.Abort()
			return
		}

		sessionActive, err := authMiddleware.DBHelper.IsUserSessionActive(sessionID)
		if err != nil {
			utils.LogError("AuthenticationMiddleware", "error verifying validity of the user session", "", err)
			utils.RespondClientErr(c, errors.New("invalid token"), http.StatusUnauthorized, "invalid token", "invalid token")
			c.Abort()
			return
		}

		// reject req is session is terminated.
		if !sessionActive {
			err := errors.New("user session is not active")
			utils.LogError("AuthenticationMiddleware", "user session is not active", "", err)
			utils.RespondClientErr(c, errors.New("invalid token"), http.StatusUnauthorized, "invalid token", "invalid token")
			c.Abort()
			return
		}

		// Now increase the time.
		err = authMiddleware.DBHelper.UpdateUserSession(sessionID)
		if err != nil {
			utils.LogError("AuthenticationMiddleware", "Updating session", "Update user session", err)
			utils.RespondClientErr(c, err, http.StatusUnauthorized, "UpdateSession: error updating sessions ", "UpdateSession error updating sessions")
			c.Abort()
			return
		}

		// Extract data values from the claims.
		// data := claims["data"].(map[string]interface{})
		issuer := claims["iss"].(string)

		userData, err := authMiddleware.DBHelper.GetUserByID(issuer)
		if err != nil {
			utils.LogError("AuthenticationMiddleware", "finding user by ID extracted from the claims", "", err)
			utils.RespondClientErr(c, err, http.StatusUnauthorized, "UpdateSession: error getting user Details", "error getting user details")
			c.Abort()
			return
		}

		// Construct the user context data.
		userContextData.ID = issuer
		userContextData.Name = userData.Name
		userContextData.Username = userData.Username
		userContextData.Quota = userData.Quota
		userContextData.UsedStorage = userData.UsedStorage

		// setting the value in the context.
		ctxWithUser := context.WithValue(c.Request.Context(), models.UserContextKey, &userContextData)
		c.Request = c.Request.WithContext(ctxWithUser)

	}
}

func GetClaimsFromToken(tokenString string) (jwt.MapClaims, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return models.JwtSigningSecretKey, nil
	})
	if err != nil {
		return jwt.MapClaims{}, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return jwt.MapClaims{}, err
}

func getUserDataFromClaims(dbHelper providers.DBHelperProvider, claims jwt.MapClaims) (string, bool, error) {

	fmt.Println("claims -  ", claims)
	data := claims["data"].(map[string]interface{})
	token := data["token"].(string)

	IsUserSessionTokenActive, err := dbHelper.IsUserSessionTokenActive(token)
	if err != nil {
		utils.LogError("getUserDataFromClaims", "error fetching user session data from database", "", err)
		logrus.Error("GetUserDataFromClaims: error fetching user session data from database ", err)
		return "", false, errors.New(fmt.Sprintln("GetUserDataFromClaims: error fetching user Data from database  & \n", err))
	}

	if IsUserSessionTokenActive {
		sessionData, err := dbHelper.ReadUserSessionBySessionToken(token)
		if err != nil {
			utils.LogError("getUserDataFromClaims", "error fetching user session data from database", "", err)
			logrus.Error("GetUserDataFromClaims: error fetching user session data from database ", err)
			return "", false, errors.New(fmt.Sprintln("GetUserDataFromClaims: error fetching user Data from database  & \n", err))
		}
		return sessionData.ID, true, nil
	}

	return "", false, errors.New(fmt.Sprintln("invalid session id or session is expired", err))
}

// Extract the user context data from the user context attached to the request.
func (authMiddleware Middleware) UserFromContext(ctx context.Context) *models.UserContext {
	return ctx.Value(models.UserContextKey).(*models.UserContext)
}
