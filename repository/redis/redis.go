package redis

import (
	"fmt"

	"github.com/g-graziano/userland/models"
	rds "github.com/go-redis/redis"
)

type redis struct {
	Redis *rds.Client
}

type Redis interface {
	// User
	CreateToken(user *models.User) error
	// UpdateUser(user *models.User) error
	// GetUser(user *models.User) ([]*models.User, error)
}

func New(conn ...string) Redis {
	client := rds.NewClient(&rds.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()

	if err != nil {
		panic(err)
	}

	fmt.Println("Redis Connected!")

	return &redis{Redis: client}
}

func (r *redis) CreateToken(user *models.User) error {
	err := r.Redis.Set("user.XID-", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	return nil
}
