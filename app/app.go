package app

import (
	_http "github.com/g-graziano/userland/delivery/http"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
}

func Run() {
	dep := buildDependency()

	_http.Router(dep.User, dep.Token)
}
