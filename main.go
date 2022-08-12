package main

import (
	"github.com/gin-gonic/gin"
	"golang-jwt-project/routes"
	"os"
)

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	router := gin.New()
	router.Use(gin.Logger())
	routes.AuthRoutes(router)
	routes.UserRoutes(router)
	// you can set your trust proxy here
	//router.SetTrustedProxies([]string{"192.168.1.2"})
	router.GET("/api-1/", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": "Access granted for api-1"})
	})
	router.GET("/api-2", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": "Access granted for api-2"})
	})
	router.Run(":" + port)
}
