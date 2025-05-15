package server

import (
	"github.com/gin-gonic/gin"
)

func (srv *Server) InjectRoutes() *gin.Engine {

	router := gin.Default()

	// Public routes
	router.POST("/login", srv.login)
	router.POST("/register", srv.createNewUser)

	// Protected routes
	protected := router.Group("/")
	protected.Use(srv.MiddlewareProvider.AuthMiddleware())
	{
		protected.GET("/storage/remaining", srv.remainingStorage)
		protected.POST("/upload", srv.uploadFile)
		protected.GET("/files", srv.getUserFiles)

	}

	return router
}
