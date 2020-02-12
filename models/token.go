package models

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type UserToken struct {
	ID        uint64    `gorm:"primary_key; AUTO_INCREMENT" json:"id"`
	UserID    uint64    `gorm:"not null" json:"user_id"`
	Token     string    `gorm:"not null" json:"token"`
	TokenType string    `gorm:"not null" json:"token_type"`
	Status    string    `gorm:"not null; type:varchar(255)" json:"status"`
	CreatedAt time.Time `gorm:"not null" json:"-"`
	UpdatedAt time.Time `gorm:"not null" json:"-"`
}

type TfaResponse struct {
	RequiredTfa *bool        `json:"required_tfa"`
	AccessToken *AccessToken `json:"access_token"`
}

type AccessToken struct {
	Value     string `json:"value"`
	Type      string `json:"type"`
	ExpiredAt string `json:"expired_at"`
}

type ForgotPassToken struct {
	Code  string `json:"code"`
	Email string `json:"email"`
}

type EmailTokenClaim struct {
	jwt.StandardClaims
	Email      string `json:"phone_number"`
	AccessType string `json:"access_type"`
}
