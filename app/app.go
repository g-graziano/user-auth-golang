package app

import (
	"context"

	_http "github.com/g-graziano/user-auth-golang/delivery/http"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
}

func Run() {
	ctx := context.Background()

	dep := buildDependency()

	_http.Router(ctx, dep.User, dep.Token)
}
