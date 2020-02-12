package user

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/g-graziano/userland/helper"
	"github.com/g-graziano/userland/models"
	"github.com/g-graziano/userland/repository/postgres"
	"github.com/g-graziano/userland/repository/redis"
	"github.com/go-playground/validator"
	"github.com/rs/xid"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"golang.org/x/crypto/bcrypt"
)

type User interface {
	Login(user *models.User) (*models.User, error)
	Logout(user *models.User) error
	Register(user *models.User) (*models.User, error)
	VerifyEmail(user *models.User) error
	GetUserProfile(user *models.User) (*models.User, error)
	SearchUserByEmail(xid string, user *models.User) (*models.User, error)
	SendEmailValidation(user *models.User) error
}

type user struct {
	postgres postgres.Postgres
	redis    redis.Redis
}

func New(pg postgres.Postgres, rd redis.Redis) User {
	return &user{
		postgres: pg,
		redis:    rd,
	}
}

func (u *user) SendEmailValidation(user *models.User) error {
	if user.Fullname == "" && user.Email != "" {
		newUser, err := u.postgres.GetUser(user)
		if err != nil {
			panic(err)
		}

		if len(newUser) < 1 {
			return errors.New("user not found")
		}

		user = newUser[0]
	}

	from := mail.NewEmail("User Land", "verifier@userland.com")
	subject := "Please verify your email address"
	to := mail.NewEmail(user.Fullname, user.Email)
	plainTextContent := "Hi " + user.Fullname + ", Please click link below to verify your email address so we know that it's really you!"
	htmlContent := `<p>Hi ` + user.Fullname + `,</p>
		<p>Please click link below to verify your email address so we know that it's really you!</p>
		<p><a href="http://0.0.0.0:8080/user/verification/` + user.XID + `" style="box-sizing: border-box;
		border-color: #ED3237;font-weight: 400;text-decoration: none;display: inline-block;margin: 0;color: #ffffff;background-color: #ED3237;
		border: solid 1px #ED3237;border-radius: 2px;font-size: 14px;padding: 12px 45px;">Confirm Email Address<a></p>`
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	_, err := client.Send(message)

	if err != nil {
		panic(err)
	}

	return nil
}

func (u *user) validate(user *models.User) error {
	validate := validator.New()

	if err := validate.Var(user.Email, "required,email,min=5,max=50"); err != nil {
		return errors.New("Email harus valid dan terdiri dari 5 s/d 50 karakter")
	}

	if err := validate.Var(user.Password, "required,min=5,max=20"); err != nil {
		return errors.New("Password harus terdiri dari 5 s/d 20 karakter")
	}

	user.Email = strings.ToLower(user.Email)
	getUser, err := u.postgres.GetUser(user)

	if err != nil {
		return errors.New("Connection Error. Please Retry")
	}

	if len(getUser) > 0 {
		return errors.New("Email Already Exists")
	}

	return nil
}

func (u *user) Login(user *models.User) (*models.User, error) {
	var getUser []*models.User

	user.Email = strings.ToLower(user.Email)

	getUser, err := u.postgres.GetUser(user)

	if len(getUser) < 1 {
		return nil, errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(getUser[0].Password), []byte(user.Password)); err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return nil, errors.New("invalid login credentials, please try again")
	}

	expireToken := time.Now().Add(time.Hour * 12).Unix()

	signKey := []byte(os.Getenv("JWT_SIGNATURE_KEY"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": getUser[0].Email,
		"xid":   getUser[0].XID,
		"exp":   expireToken,
	})

	tokenString, _ := token.SignedString(signKey)
	getUser[0].Token = helper.NullStringFunc(tokenString, true)

	err = u.postgres.UpdateUser(getUser[0])
	if err != nil {
		panic(err)
	}

	getUser[0].Password = ""

	return getUser[0], err
}

func (u *user) Logout(user *models.User) error {
	var newUser, err = u.postgres.GetUser(&models.User{XID: user.XID})

	if len(newUser) < 1 {
		return errors.New("user not found")
	}

	newUser[0].Token = helper.NullStringFunc("", false)

	err = u.postgres.UpdateUser(newUser[0])
	if err != nil {
		panic(err)
	}
	return nil
}

func (u *user) Register(user *models.User) (*models.User, error) {
	if err := u.validate(user); err != nil {
		return nil, err
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	user.Password = string(hashedPassword)
	user.Email = strings.ToLower(user.Email)
	user.XID = xid.New().String()

	var err = u.postgres.CreateUser(user)
	if err != nil {
		panic(err)
	}

	newUser, err := u.postgres.GetUser(&models.User{XID: user.XID})
	if err != nil {
		panic(err)
	}

	if len(newUser) < 1 {
		return nil, errors.New("Register failed")
	}

	u.SendEmailValidation(user)

	return user, err
}

func (u *user) VerifyEmail(user *models.User) error {
	var verifyUser, err = u.postgres.GetUser(user)

	if err != nil {
		panic(err)
	}

	if len(verifyUser) < 1 {
		return errors.New("user not found")
	}

	verifyUser[0].Status = "verified"

	err = u.postgres.UpdateUser(verifyUser[0])
	if err != nil {
		panic(err)
	}

	return nil
}

func (u *user) SearchUserByEmail(xid string, user *models.User) (*models.User, error) {
	var getUser, err = u.postgres.GetUser(&models.User{XID: xid})

	if err != nil {
		panic(err)
	}

	if getUser[0].Email == user.Email {
		return nil, errors.New("user not found")
	}

	searchUser, err := u.postgres.GetUser(user)

	if err != nil {
		panic(err)
	}

	if len(searchUser) < 1 {
		return nil, errors.New("user not found")
	}

	searchUser[0].Password = ""
	searchUser[0].Token = helper.NullStringFunc("", false)

	return searchUser[0], nil
}

func (u *user) GetUserProfile(user *models.User) (*models.User, error) {
	var foundUser, err = u.postgres.GetUser(user)

	if err != nil {
		panic(err)
	}

	if len(foundUser) < 1 {
		return nil, errors.New("user not found")
	}

	foundUser[0].Password = ""
	foundUser[0].Token = helper.NullStringFunc("", false)

	return foundUser[0], nil
}
