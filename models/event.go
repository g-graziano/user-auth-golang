package models

import (
	"time"

	"github.com/g-graziano/userland/helper"
)

type Event struct {
	ID         uint64            `gorm:"primary_key; AUTO_INCREMENT" json:"-"`
	UserID     uint64            `gorm:"not null" json:"user_id"`
	Event      string            `gorm:"null" json:"event"`
	UA         string            `gorm:"null" json:"ua"`
	IPAddress  helper.NullString `gorm:"null" json:"ip_address"`
	ClientID   uint64            `gorm:"null" json:"client_id"`
	CreatedAt  time.Time         `gorm:"not null" json:"created_at"`
	UpdatedAt  time.Time         `gorm:"not null" json:"-"`
	ClientName string            `gorm:"-" json:"client_name"`
	TotalEvent int               `gorm:"-" json:"total_event"`
}

type ListEventResponse struct {
	Data       []ListEventResponseData `json:"data"`
	Pagination Pagination              `json:"pagination"`
}

type ListEventResponseData struct {
	Event     string            `json:"event"`
	UA        string            `json:"ua"`
	IP        helper.NullString `json:"ip"`
	Client    DataClient        `json:"client"`
	CreatedAt time.Time         `json:"created_at"`
}

type Pagination struct {
	Page     int `json:"page"`
	PerPage  int `json:"per_page"`
	Next     int `json:"next"`
	Previous int `json:"previous"`
}
