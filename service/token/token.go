package token

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/g-graziano/userland/models"
	"github.com/g-graziano/userland/repository/postgres"
	"github.com/g-graziano/userland/repository/redis"
)

type Token interface {
	VerifyTfa(otp *models.OTPRequest) (*models.AccessToken, error)
}

type token struct {
	postgres postgres.Postgres
	redis    redis.Redis
}

func New(pg postgres.Postgres, rd redis.Redis) Token {
	return &token{
		postgres: pg,
		redis:    rd,
	}
}

func (t *token) VerifyTfa(otp *models.OTPRequest) (*models.AccessToken, error) {
	verifyUser, err := t.postgres.GetUser(&models.User{XID: otp.XID})

	if err != nil {
		return nil, err
	}

	_, err = t.redis.Get(&models.OTP{Key: strconv.FormatUint(verifyUser[0].ID, 10) + "-login"})

	if err != nil {
		return nil, errors.New("OTP tidak berlaku")
	}

	now := time.Now().UTC()
	end := now.Add(time.Hour * 24)
	claim := models.TokenClaim{
		XID:        verifyUser[0].XID,
		Email:      verifyUser[0].Email,
		AccessType: "login",
	}
	claim.IssuedAt = now.Unix()
	claim.ExpiresAt = end.Unix()

	signKey := []byte(os.Getenv("JWT_SIGNATURE_KEY"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	tokenString, _ := token.SignedString(signKey)

	err = t.postgres.CreateToken(&models.UserToken{Token: tokenString, UserID: verifyUser[0].ID, TokenType: "login"})

	if err != nil {
		return nil, err
	}

	return &models.AccessToken{
		Value:     tokenString,
		Type:      "Bearer",
		ExpiredAt: end.String(),
	}, nil
}
