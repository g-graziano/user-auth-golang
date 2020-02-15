package models

import (
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/g-graziano/userland/helper"
)

type UserToken struct {
	ID           uint64            `gorm:"primary_key; AUTO_INCREMENT" json:"id"`
	UserID       uint64            `gorm:"not null" json:"user_id"`
	Token        string            `gorm:"not null" json:"token"`
	TokenType    string            `gorm:"not null" json:"token_type"`
	RefreshToken helper.NullString `gorm:"null" json:"refresh_token"`
	Status       string            `gorm:"not null; type:varchar(255)" json:"status"`
	IPAddress    string            `gorm:"null" json:"ip_address"`
	ClientID     uint64            `gorm:"null" json:"client_id"`
	CreatedAt    time.Time         `gorm:"not null" json:"-"`
	UpdatedAt    time.Time         `gorm:"not null" json:"-"`
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

type AccessTokenRequest struct {
	XID          string `json:"value"`
	RefreshToken string `json:"type"`
}

type ForgotPassToken struct {
	Code  string `json:"code"`
	Email string `json:"email"`
}

type TokenClaim struct {
	jwt.StandardClaims
	XID        string        `json:"xid"`
	Email      string        `json:"phone_number"`
	AccessType string        `json:"access_type"`
	ExpiredAt  time.Duration `json:"expired_at"`
}

func (e *TokenClaim) TokenGenerator() string {
	now := time.Now().UTC()
	end := now.Add(e.ExpiredAt)

	var claim TokenClaim

	if e.XID != "" {
		claim = TokenClaim{
			XID:        e.XID,
			Email:      e.Email,
			AccessType: e.AccessType,
		}
	} else {
		claim = TokenClaim{
			Email:      e.Email,
			AccessType: e.AccessType,
		}
	}

	claim.IssuedAt = now.Unix()
	claim.ExpiresAt = end.Unix()

	signKey := []byte(os.Getenv("JWT_SIGNATURE_KEY"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	tokenString, _ := token.SignedString(signKey)

	return tokenString
}

func VerifyToken(tokenString string) (*TokenClaim, error) {
	signKey := []byte(os.Getenv("JWT_SIGNATURE_KEY"))

	claim := new(TokenClaim)

	_, err := jwt.ParseWithClaims(tokenString, claim, func(token *jwt.Token) (interface{}, error) {
		return []byte(signKey), nil
	})

	if err != nil {
		return nil, err
	}

	return claim, nil
}
