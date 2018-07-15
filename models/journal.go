package models

import "time"

type Journal struct {
	ID            string    `gorm:"primary_key" json:"id"`
	FromAccountID string    `json:"from_account_id"`
	ToAccountID   string    `json:"to_account_id"`
	Amount        int64     `json:"balance"`
	Charge        int64     `json:"charge"`
	Status        int       `json:"status"` // 1:normal 2:processing 3:fail
	CreatedAt     time.Time `json:"created_at"`
}
