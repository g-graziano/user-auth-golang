package models

type ClientID struct {
	ID   uint64 `gorm:"primary_key; AUTO_INCREMENT" json:"id"`
	API  string `gorm:"not null" json:"api"`
	Name string `gorm:"not null" json:"name"`
}
