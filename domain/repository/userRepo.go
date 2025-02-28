package repository

import (
	"github.com/jmoiron/sqlx"
	"wallet-service/domain/entity"
)

type PostgresUserRepository struct {
	db *sqlx.DB
}

func NewPostgresUserRepository(db *sqlx.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) InitializeSchema() error {
	_, err := r.db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, username VARCHAR(255) NOT NULL UNIQUE, password VARCHAR(255) NOT NULL)")
	return err
}

func (r *PostgresUserRepository) UserRegistration(user entity.User) error {
	// Вставка пользователя в таблицу users
	err := r.db.QueryRow(
		"INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id",
		user.Username, user.Password,
	).Scan(&user.ID)

	if err != nil {
		return err
	}

	// Вставка записи в таблицу wallets с использованием полученного user.ID
	_, err = r.db.Exec("INSERT INTO wallets (user_id, balance) VALUES ($1, 0)", user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresUserRepository) GetUserHashedPass(user entity.User) (string, error) {
	var storedPassword string
	err := r.db.QueryRow("SELECT password FROM users WHERE username=$1", user.Username).Scan(&storedPassword)

	return storedPassword, err
}

func (r *PostgresUserRepository) GetUserID(user entity.User) (int, error) {
	var ID int
	err := r.db.QueryRow("SELECT id FROM users WHERE username=$1", user.Username).Scan(&ID)
	if err != nil {
		return 0, err
	}
	return ID, nil
}
