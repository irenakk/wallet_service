package main

type User struct {
	ID       int    `db:"id"`
	Username string `db:"username"`
}

type Wallet struct {
	ID      int     `db:"id"`
	UserID  int     `db:"user_id"`
	Balance float64 `db:"balance"`
}
