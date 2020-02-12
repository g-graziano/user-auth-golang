package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"

	"github.com/g-graziano/userland/helper"
)

func VerifyToken(tokenString string) (jwt.Claims, error) {
	signKey := []byte(os.Getenv("JWT_SIGNATURE_KEY"))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (i interface{}, err error) {
		return signKey, err
	})

	if err != nil {
		return nil, err
	}

	return token.Claims, nil
}

func JwtAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := make(map[string]interface{})
		tokenString := r.Header.Get("Authorization")

		if len(tokenString) == 0 {
			response = helper.Message(false, "Missing auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json; charset=utf-8")
			helper.Response(w, response)
			return
		}

		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
		claims, err := VerifyToken(tokenString)
		if err != nil {
			response = helper.Message(false, "Invalid auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json; charset=utf-8")
			helper.Response(w, response)
			return
		}

		xid := claims.(jwt.MapClaims)["xid"].(string)
		if jwtType := claims.(jwt.MapClaims)["type"].(string); jwtType != "login" {
			response = helper.Message(false, "Invalid auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json; charset=utf-8")
			helper.Response(w, response)
			return
		}

		r.Header.Set("xid", xid)
		r.Header.Set("token", tokenString)

		next.ServeHTTP(w, r)
	})
}

func JwtTfaAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := make(map[string]interface{})
		tokenString := r.Header.Get("Authorization")

		if len(tokenString) == 0 {
			response = helper.Message(false, "Missing auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json; charset=utf-8")
			helper.Response(w, response)
			return
		}

		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
		claims, err := VerifyToken(tokenString)
		if err != nil {
			response = helper.Message(false, "Invalid auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json; charset=utf-8")
			helper.Response(w, response)
			return
		}

		email := claims.(jwt.MapClaims)["email"].(string)
		xid := claims.(jwt.MapClaims)["xid"].(string)

		r.Header.Set("email", email)
		r.Header.Set("xid", xid)
		r.Header.Set("token", tokenString)

		next.ServeHTTP(w, r)
	})
}
