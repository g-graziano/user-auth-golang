package middleware

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"

	"github.com/g-graziano/userland/helper"
	"github.com/g-graziano/userland/models"
	"github.com/g-graziano/userland/service/user"
)

func VerifyToken(tokenString string) (*models.TokenClaim, error) {
	if len(tokenString) == 0 {
		return nil, errors.New("Missing auth token")
	}

	tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

	signKey := []byte(os.Getenv("JWT_SIGNATURE_KEY"))

	claim := new(models.TokenClaim)

	_, err := jwt.ParseWithClaims(tokenString, claim, func(token *jwt.Token) (interface{}, error) {
		return []byte(signKey), nil
	})

	if err != nil {
		return nil, err
	}

	return claim, nil
}

func APIClientAuthentication(u user.User) (ret func(http.Handler) http.Handler) {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("X-API-ClientID")

			client, err := u.GetAPIClientID(&models.ClientID{API: tokenString})

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				helper.Response(w, helper.ErrorMessage(0, err.Error()))

				return
			}

			if client.API == "" {
				w.WriteHeader(http.StatusBadRequest)
				helper.Response(w, helper.ErrorMessage(0, "Client API not valid"))

				return
			}

			r.Header.Set("client-id", strconv.FormatUint(client.ID, 10))

			next.ServeHTTP(w, r)
		})
	}
}

func JwtTfaAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")

		claims, err := VerifyToken(r.Header.Get("Authorization"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))

			return
		}

		if jwtType := claims.AccessType; jwtType != "tfa" {
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json; charset=utf-8")
			helper.Response(w, helper.ErrorMessage(0, "Invalid auth token"))
			return
		}

		r.Header.Set("xid", claims.XID)
		r.Header.Set("token", tokenString)
		r.Header.Set("client-id", strconv.FormatUint(claims.ClientID, 10))

		next.ServeHTTP(w, r)
	})
}

func JwtACTAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")

		claims, err := VerifyToken(r.Header.Get("Authorization"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))

			return
		}

		if jwtType := claims.AccessType; jwtType != "refreshtoken" {
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json; charset=utf-8")
			helper.Response(w, helper.ErrorMessage(0, "Invalid auth token"))
			return
		}

		r.Header.Set("xid", claims.XID)
		r.Header.Set("token", tokenString)
		r.Header.Set("client-id", strconv.FormatUint(claims.ClientID, 10))

		next.ServeHTTP(w, r)
	})
}

func JwtAuthentication(u user.User) (ret func(http.Handler) http.Handler) {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")

			claims, err := VerifyToken(r.Header.Get("Authorization"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				helper.Response(w, helper.ErrorMessage(0, err.Error()))

				return
			}

			if jwtType := claims.AccessType; jwtType != "login" {
				w.WriteHeader(http.StatusForbidden)
				w.Header().Add("Content-Type", "application/json; charset=utf-8")
				helper.Response(w, helper.ErrorMessage(0, "Invalid auth token"))
				return
			}

			tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
			err = u.CheckJWTIsActive(&models.UserToken{Token: tokenString})

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				helper.Response(w, helper.ErrorMessage(0, err.Error()))
				return
			}

			r.Header.Set("email", claims.Email)
			r.Header.Set("xid", claims.XID)
			r.Header.Set("token", tokenString)
			r.Header.Set("client-id", strconv.FormatUint(claims.ClientID, 10))

			next.ServeHTTP(w, r)
		})
	}
}
