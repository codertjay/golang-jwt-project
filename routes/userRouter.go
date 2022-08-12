package routes

import (
	"github.com/gin-gonic/gin"
	controller "golang-jwt-project/controllers"
	"golang-jwt-project/middleware"
)

func userRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticated())
	incomingRoutes.GET("/users", controller.GetUsers())
	incomingRoutes.GET("/users/:user_id", controller.GetUser())

}
