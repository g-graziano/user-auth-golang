package models

import (
	"io"
	"time"

	"github.com/g-graziano/userland/helper"
)

type User struct {
	ID       uint64 `gorm:"primary_key; AUTO_INCREMENT" json:"id"`
	XID      string `gorm:"unique" json:"xid"`
	Email    string `gorm:"unique; not null; type:varchar(255)" json:"email"`
	Fullname string `gorm:"not null; type:varchar(255)" json:"fullname"`

	Location helper.NullString `gorm:"type:varchar(255)" json:"location"`
	Bio      helper.NullString `gorm:"type:varchar(255)" json:"bio"`
	Web      helper.NullString `gorm:"type:varchar(255)" json:"web"`
	Picture  helper.NullString `gorm:"type:varchar(255)" json:"picture,omitempty"`

	Password  string            `gorm:"type:varchar(255); not null" json:"password,omitempty"`
	Token     helper.NullString `gorm:"type:varchar(255)" json:"token,omitempty"`
	Status    int               `gorm:"not null; default: 1" json:"-"`
	CreatedAt time.Time         `gorm:"not null" json:"-"`
	UpdatedAt time.Time         `gorm:"not null" json:"-"`
}

type ProfileResponse struct {
	ID        uint64    `json:"id"`
	Fullname  string    `json:"fullname"`
	Location  string    `json:"location"`
	Bio       string    `json:"bio"`
	Web       string    `json:"web"`
	Picture   string    `json:"picture"`
	CreatedAt time.Time `json:"created_at"`
}

type GetEmailResponse struct {
	Email string `json:"email"`
}

type ChangePassword struct {
	XID             string `json:"xid"`
	PasswordCurrent string `json:"password_current"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

type SetPictureRequest struct {
	CourierID   int64     `json:"-"`
	Size        int64     `json:"-"`
	ContentType string    `json:"-"`
	Width       int64     `json:"-"`
	Height      int64     `json:"-"`
	File        io.Reader `json:"-"`
}

type RegisterRequest struct {
	Token           string `json:"token"`
	Fullname        string `json:"fullname"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

type VerificationRequest struct {
	Type      string `json:"type"`
	Recipient string `json:"recipient"`
}

type ResetPass struct {
	Token           string `json:"token"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}
