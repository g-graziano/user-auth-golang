package app

import (
	"fmt"
	"os"

	"github.com/g-graziano/user-auth-golang/repository/postgres"
	"github.com/g-graziano/user-auth-golang/repository/redis"
	"github.com/g-graziano/user-auth-golang/service/token"
	"github.com/g-graziano/user-auth-golang/service/user"
)

type Dependency struct {
	User  user.User
	Token token.Token
	// Point        point.Point
	// PointHistory pointHistory.PointHistory
}

func buildDependency() Dependency {
	var dep Dependency
	dbHost := os.Getenv("DATABASE_HOST")
	dbPort := os.Getenv("DATABASE_PORT")
	dbUser := os.Getenv("DATABASE_USER")
	dbPass := os.Getenv("DATABASE_PASS")
	dbName := os.Getenv("DATABASE_NAME")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)
	pg := postgres.New(connStr)
	rd := redis.New()

	dep.User = user.New(pg, rd)
	dep.Token = token.New(pg, rd)
	return dep
}
