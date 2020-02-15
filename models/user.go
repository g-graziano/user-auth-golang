package models

import (
	"errors"
	"io"
	"strings"
	"time"

	"github.com/g-graziano/userland/helper"
	"github.com/go-playground/validator"
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

	Password     string          `gorm:"type:varchar(255); not null" json:"password,omitempty"`
	Status       string          `gorm:"not null; default: 'nonactive'" json:"-"`
	TFA          bool            `gorm:"type:varchar(255); not null; default: false" json:"tfa"`
	EnabledTfaAt helper.NullTime `gorm:"null" json:"enabled_tfa_at"`

	TFASecret helper.NullTime `gorm:"-" json:"tfa_secret"`
	TFAQr     helper.NullTime `gorm:"-" json:"tfa_qr"`

	IPAddress string `gorm:"-" json:"-"`

	CreatedAt time.Time `gorm:"not null" json:"-"`
	UpdatedAt time.Time `gorm:"not null" json:"-"`
}

type Login struct {
	XID       string `json:"xid"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	IPAddress string `json:"ip_address"`
}

type UploadProfile struct {
	UserXID     string    `json:"-"`
	Size        int64     `json:"-"`
	ContentType string    `json:"-"`
	Width       int64     `json:"-"`
	Height      int64     `json:"-"`
	File        io.Reader `json:"-"`
}

type EnrollTfa struct {
	Secret string `json:"secret"`
	Qr     string `json:"qr"`
}

type ActivateTfaRequest struct {
	XID    string `json:"-"`
	Secret string `json:"secret"`
	Code   string `json:"code"`
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

type ChangePassword struct {
	XID             string `json:"xid"`
	PasswordCurrent string `json:"password_current"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

type ResetPass struct {
	Token           string `json:"token"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

type TFAStatus struct {
	Enabled   bool      `json:"enabled"`
	EnabledAt time.Time `json:"enabled_at"`
}

type Email struct {
	RecipientName  string
	RecipientEmail string
	Subject        string
	HTMLContent    string
	PlainContent   string
}

func (user *RegisterRequest) ValidateRegister() error {
	validate := validator.New()

	if err := validate.Var(user.Email, "required,email,min=5,max=50"); err != nil {
		return errors.New("Email harus valid dan terdiri dari 5 s/d 50 karakter")
	}

	if err := validate.Var(user.Password, "required,min=5,max=20"); err != nil {
		return errors.New("Password harus terdiri dari 5 s/d 20 karakter")
	}

	if err := validate.VarWithValue(user.Password, user.PasswordConfirm, "eqfield"); err != nil {
		return errors.New("Password tidak sama")
	}

	user.Email = strings.ToLower(user.Email)

	return nil
}
