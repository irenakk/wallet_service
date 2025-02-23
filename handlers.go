package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func deposit(c *gin.Context) {
	var req struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "UPDATE wallets SET balance = balance + $1 WHERE user_id = $2"
	_, err := db.Exec(query, req.Amount, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка пополнения баланса"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Баланс пополнен"})
}

func transfer(c *gin.Context) {
	var req struct {
		FromUserID int     `json:"from_user_id"`
		ToUserID   int     `json:"to_user_id"`
		Amount     float64 `json:"amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка транзакции"})
		return
	}

	_, err = tx.Exec("UPDATE wallets SET balance = balance - $1 WHERE user_id = $2 AND balance >= $1", req.Amount, req.FromUserID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Недостаточно средств"})
		return
	}

	_, err = tx.Exec("UPDATE wallets SET balance = balance + $1 WHERE user_id = $2", req.Amount, req.ToUserID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка перевода"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Перевод выполнен"})
}

func registerUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "INSERT INTO users (username) VALUES ($1) RETURNING id"
	err := db.QueryRow(query, user.Username).Scan(&user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка регистрации"})
		fmt.Println(err)
		return
	}

	query = "INSERT INTO wallets (user_id, balance) VALUES ($1, 0)"
	_, err = db.Exec(query, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания кошелька"})
		return
	}

	c.JSON(http.StatusOK, user)
}
