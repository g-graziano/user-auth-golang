package models

import (
	"time"

	"github.com/g-graziano/userland/helper"
)

type User struct {
	ID       uint64 `gorm:"primary_key; AUTO_INCREMENT" json:"id"`
	XID      string `gorm:"unique" json:"xid"`
	Email    string `gorm:"unique; not null; type:varchar(255)" json:"email"`
	Fullname string `gorm:"not null; type:varchar(255)" json:"name"`

	Location helper.NullString `gorm:"type:varchar(255)" json:"location"`
	Bio      helper.NullString `gorm:"type:varchar(255)" json:"bio"`
	Web      helper.NullString `gorm:"type:varchar(255)" json:"web"`
	Picture  helper.NullString `gorm:"type:varchar(255)" json:"picture"`

	Password  string            `gorm:"type:varchar(255); not null" json:"password"`
	Token     helper.NullString `gorm:"type:varchar(255)" json:"token,omitempty"`
	Status    string            `gorm:"not null; default:'unverifyed'" json:"-"`
	CreatedAt time.Time         `gorm:"not null" json:"-"`
	UpdatedAt time.Time         `gorm:"not null" json:"-"`
}
