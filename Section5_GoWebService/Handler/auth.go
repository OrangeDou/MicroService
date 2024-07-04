package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleWare() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Request.Header.Get("Authorization")

		authorized := check(token)
		if authorized {
			ctx.Next()
			return
		}

		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		ctx.Abort()
		return
	}
}

//最好分开写，service方法

func check(token string) bool {
	if token == "admin" {
		return true
	}
	return false
}
