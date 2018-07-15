package models

type Account struct {
	ID      string `gorm:"primary_key" json:"id"`
	OwnerId string `json:"owner_id"`
	Balance int64  `json:"balance"`
	Status  int    `json:"status"` // 1:normal 2:deleted
}
