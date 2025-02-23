package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	initDB()

	r := gin.Default()
	r.POST("/register", registerUser)
	r.POST("/deposit", deposit)
	r.POST("/transfer", transfer)

	err := r.Run(":8083")
	if err != nil {
		return
	}
}
