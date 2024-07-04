package main

import (
	handler "micro/Section5_GoWebService/Handler"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.GET("/user:name", func(c *gin.Context) {
		name := c.Param("name")
		c.String(http.StatusOK, "Hello, %s", name)
	})

	router.GET("/path", handler.AuthMiddleWare(), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"data": "ok",
		})
	})

	router.Run(":8080")
}
