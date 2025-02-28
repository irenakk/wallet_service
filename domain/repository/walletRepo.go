package repository

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
)

type PostgresWalletRepository struct {
	db *sqlx.DB
}

func NewWalletRepository(db *sqlx.DB) *PostgresWalletRepository {
	return &PostgresWalletRepository{db: db}
}

func (r *PostgresWalletRepository) InitializeSchema() error {
	_, err := r.db.Exec("CREATE TABLE IF NOT EXISTS wallets (id SERIAL PRIMARY KEY, user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE, balance NUMERIC NOT NULL DEFAULT 0)")
	return err
}

func (r *PostgresWalletRepository) Deposit(userID int, amount float64) error {
	_, err := r.db.Exec("UPDATE wallets SET balance = balance + $1 WHERE user_id = $2", amount, userID)
	if err != nil {
		return err
	}
	return err
}

func (r *PostgresWalletRepository) Transfer(fromUserID, toUserID int, amount float64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// Вычитаем сумму у отправителя, проверяя, что средств достаточно
	res, err := tx.Exec("UPDATE wallets SET balance = balance - $1 WHERE user_id = $2 AND balance >= $1", amount, fromUserID)
	if err != nil {
		tx.Rollback()
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		tx.Rollback()
		return sql.ErrNoRows // Ошибка, если средств недостаточно
	}

	// Добавляем сумму получателю
	_, err = tx.Exec("UPDATE wallets SET balance = balance + $1 WHERE user_id = $2", amount, toUserID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
