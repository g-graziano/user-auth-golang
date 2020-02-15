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
	Create(otp *models.OTP) error
	Get(otp *models.OTP) (string, error)
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

func (r *redis) Create(otp *models.OTP) error {
	err := r.Redis.Set(otp.Key, otp.Value, 0).Err()

	r.Redis.ExpireAt(otp.Key, otp.Expire)

	if err != nil {
		return err
	}

	return nil
}

func (r *redis) Get(otp *models.OTP) (string, error) {
	res, err := r.Redis.Get(otp.Key).Result()
	if err != nil {
		return "", err
	}

	return res, nil
}
