package user

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/g-graziano/userland/helper"
	"github.com/g-graziano/userland/models"
	"github.com/g-graziano/userland/repository/postgres"
	"github.com/g-graziano/userland/repository/redis"
	sdg "github.com/g-graziano/userland/repository/sendgrid"
	"github.com/go-playground/validator"
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

type User interface {
	Login(user *models.User) (*models.AccessToken, error)
	Logout(user *models.User) error
	Register(user *models.RegisterRequest) error
	VerifyEmail(user *models.User) error
	SendEmailValidation(user *models.User) error
	ResendEmailValidation(user *models.User) error
	ForgotPassword(user *models.User) error
	ResetPassword(resetPass *models.ResetPass) error

	GetUserProfile(user *models.User) (*models.ProfileResponse, error)
	GetUserEmail(user *models.User) (*models.GetEmailResponse, error)
	UpdateUserProfile(user *models.User) error
	UpdateUserPassword(user *models.ChangePassword) error
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
			return err
		}

		if len(newUser) < 1 {
			return errors.New("user not found")
		}

		user = newUser[0]
	}

	var email models.Email
	email.Subject = "Please verify your email address"
	email.RecipientName = user.Fullname
	email.RecipientEmail = user.Email
	email.PlainContent = "Hi " + user.Fullname + ", Please click link below to verify your email address so we know that it's really you!"
	email.HTMLContent = `<p>Hi ` + user.Fullname + `,</p>
		<p>Please click link below to verify your email address so we know that it's really you!</p>
		<p><a href="http://0.0.0.0:8080/auth/verification/` + user.XID + `" style="box-sizing: border-box;
		border-color: #ED3237;font-weight: 400;text-decoration: none;display: inline-block;margin: 0;color: #ffffff;background-color: #ED3237;
		border: solid 1px #ED3237;border-radius: 2px;font-size: 14px;padding: 12px 45px;">Confirm Email Address<a></p>`

	err := sdg.SendEmail(&email)

	if err != nil {
		return err
	}

	return nil
}

func (u *user) SendEmailOTP(user *models.User) error {
	OTP := helper.GenerateOTP()

	err := u.redis.Create(&models.OTP{Value: OTP, Key: strconv.FormatUint(user.ID, 10) + "-login", Expire: time.Now().Add(time.Minute * 5)})
	if err != nil {
		return err
	}

	var email models.Email
	email.Subject = "User land OTP"
	email.RecipientName = user.Fullname
	email.RecipientEmail = user.Email
	email.PlainContent = "Hi " + user.Fullname + ", Please input OTP below"
	email.HTMLContent = `<p>Hi ` + user.Fullname + `,</p>
		<p>Please input OTP below, so we know that it's really you!</p>
		<p style="box-sizing: border-box;
		border-color: #ED3237;font-weight: 400;text-decoration: none;display: inline-block;margin: 0;color: #ffffff;background-color: #ED3237;
		border: solid 1px #ED3237;border-radius: 2px;font-size: 14px;padding: 12px 45px;">` + OTP + `</p>`

	err = sdg.SendEmail(&email)

	if err != nil {
		return err
	}

	return nil
}

func (u *user) validateRegister(user *models.RegisterRequest) error {
	validate := validator.New()

	if err := validate.Var(user.Email, "required,email,min=5,max=50"); err != nil {
		return errors.New("Email harus valid dan terdiri dari 5 s/d 50 karakter")
	}

	if err := validate.Var(user.Password, "required,min=5,max=20"); err != nil {
		return errors.New("Password harus terdiri dari 5 s/d 20 karakter")
	}

	if err := validate.VarWithValue(user.Password, user.PasswordConfirm, "eqfield"); err != nil {
		return errors.New("Password tidak sama")
	}

	user.Email = strings.ToLower(user.Email)

	getUser, err := u.postgres.GetUser(&models.User{Email: user.Email})

	if err != nil {
		return errors.New("Connection Error. Please Retry")
	}

	if len(getUser) > 0 {
		return errors.New("Email Already Exists")
	}

	return nil
}

func (u *user) Login(user *models.User) (*models.AccessToken, error) {
	var getUser []*models.User

	user.Email = strings.ToLower(user.Email)

	getUser, err := u.postgres.GetUser(user)

	if len(getUser) < 1 {
		return nil, errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(getUser[0].Password), []byte(user.Password)); err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return nil, errors.New("invalid login credentials, please try again")
	}

	expireToken := time.Now().Add(time.Minute * 5).Unix()
	ExpiredAt := time.Now().Add(time.Minute * 5)

	signKey := []byte(os.Getenv("JWT_SIGNATURE_KEY"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"xid":  getUser[0].XID,
		"type": "login",
		"exp":  expireToken,
	})

	tokenString, _ := token.SignedString(signKey)
	getUser[0].Token = helper.NullStringFunc(tokenString, true)

	err = u.postgres.UpdateUser(getUser[0])
	if err != nil {
		return nil, err
	}

	err = u.postgres.CreateToken(&models.UserToken{Token: tokenString, UserID: getUser[0].ID, TokenType: "login"})

	accessToken := &models.AccessToken{
		Value:     tokenString,
		Type:      "bearer",
		ExpiredAt: ExpiredAt.String(),
	}

	u.SendEmailOTP(getUser[0])

	return accessToken, err
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

func (u *user) Register(user *models.RegisterRequest) error {
	if err := u.validateRegister(user); err != nil {
		return err
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	var createUser models.User

	createUser.Password = string(hashedPassword)
	createUser.Email = strings.ToLower(user.Email)
	createUser.XID = xid.New().String()

	var err = u.postgres.CreateUser(&createUser)
	if err != nil {
		return err
	}

	newUser, err := u.postgres.GetUser(&models.User{XID: createUser.XID})
	if err != nil {
		return err
	}

	if len(newUser) < 1 {
		return errors.New("Register failed")
	}

	u.SendEmailValidation(newUser[0])

	return err
}

func (u *user) VerifyEmail(user *models.User) error {
	var verifyUser, err = u.postgres.GetUser(user)

	if err != nil {
		return err
	}

	if len(verifyUser) < 1 {
		return errors.New("user not found")
	}

	verifyUser[0].Status = 2

	err = u.postgres.UpdateUser(verifyUser[0])
	if err != nil {
		return err
	}

	return nil
}

func (u *user) ResendEmailValidation(user *models.User) error {
	findUser, err := u.postgres.GetUser(&models.User{Email: user.Email})
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	if len(findUser) < 1 {
		return errors.New("user not found")
	}

	u.SendEmailValidation(findUser[0])

	return nil
}

func (u *user) ForgotPassword(user *models.User) error {
	var getUser, err = u.postgres.GetActiveUser(user)

	if err != nil {
		return err
	}

	if len(getUser) < 1 {
		return errors.New("user not found")
	}

	token := TokenGenerator(&models.EmailTokenClaim{
		Email:      getUser[0].Email,
		AccessType: "resetpassword",
	})

	var email models.Email
	email.Subject = "Reset Password"
	email.RecipientName = getUser[0].Fullname
	email.RecipientEmail = getUser[0].Email
	email.PlainContent = "Hi " + getUser[0].Fullname + ", Please input token below"
	email.HTMLContent = `<p>Hi ` + getUser[0].Fullname + `,</p>
		<p>Please input token below</p>
		<p style="box-sizing: border-box;
		border-color: #ED3237;font-weight: 400;text-decoration: none;display: inline-block;margin: 0;color: #ffffff;background-color: #ED3237;
		border: solid 1px #ED3237;border-radius: 2px;font-size: 14px;padding: 12px 45px;">` + token + `</p>`

	err = sdg.SendEmail(&email)
	if err != nil {
		return err
	}

	return nil
}

func (u *user) ResetPassword(resetPass *models.ResetPass) error {
	token, err := VerifyToken(resetPass.Token)

	if err != nil {
		return err
	}

	validate := validator.New()

	if err := validate.VarWithValue(resetPass.Password, resetPass.PasswordConfirm, "eqfield"); err != nil {
		return errors.New("Password tidak sama")
	}

	if err != nil {
		return err
	}

	if token.AccessType != "resetpassword" {
		return errors.New("token tidak valid")
	}

	getUser, err := u.postgres.GetActiveUser(&models.User{Email: token.Email})

	if err != nil {
		return err
	}

	if len(getUser) < 1 {
		return errors.New("user not found")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(resetPass.Password), bcrypt.DefaultCost)

	getUser[0].Password = string(hashedPassword)

	err = u.postgres.UpdateUser(getUser[0])
	if err != nil {
		return err
	}

	return nil
}

func (u *user) GetUserProfile(user *models.User) (*models.ProfileResponse, error) {
	var foundUser, err = u.postgres.GetUser(user)

	if err != nil {
		return nil, err
	}

	if len(foundUser) < 1 {
		return nil, errors.New("user not found")
	}

	var result models.ProfileResponse

	result.ID = foundUser[0].ID
	result.Fullname = foundUser[0].Fullname
	result.Location = foundUser[0].Location.String
	result.Bio = foundUser[0].Bio.String
	result.Web = foundUser[0].Web.String
	result.Picture = foundUser[0].Picture.String
	result.CreatedAt = foundUser[0].CreatedAt

	return &result, nil
}

func (u *user) GetUserEmail(user *models.User) (*models.GetEmailResponse, error) {
	var foundUser, err = u.postgres.GetUser(user)

	if err != nil {
		return nil, err
	}

	if len(foundUser) < 1 {
		return nil, errors.New("user not found")
	}

	var result models.GetEmailResponse
	result.Email = foundUser[0].Email

	return &result, nil
}

func (u *user) UpdateUserProfile(user *models.User) error {
	var foundUser, err = u.postgres.GetUser(user)

	if err != nil {
		return err
	}

	if len(foundUser) < 1 {
		return errors.New("user not found")
	}

	foundUser[0].Fullname = user.Fullname
	foundUser[0].Location = user.Location
	foundUser[0].Bio = user.Bio
	foundUser[0].Web = user.Web

	err = u.postgres.UpdateUser(foundUser[0])
	if err != nil {
		return err
	}

	return nil
}

func (u *user) UpdateUserPassword(user *models.ChangePassword) error {
	var foundUser, err = u.postgres.GetUser(&models.User{XID: user.XID})

	if err != nil {
		return err
	}

	if len(foundUser) < 1 {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser[0].Password), []byte(user.PasswordCurrent)); err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return errors.New("current password not valid")
	}

	validate := validator.New()
	if err := validate.VarWithValue(user.Password, user.PasswordConfirm, "eqfield"); err != nil {
		return errors.New("Password tidak sama")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	foundUser[0].Password = string(hashedPassword)

	err = u.postgres.UpdateUser(foundUser[0])
	if err != nil {
		return err
	}

	return nil
}

func TokenGenerator(e *models.EmailTokenClaim) string {
	now := time.Now().UTC()
	end := now.Add(time.Minute * 5)
	claim := models.EmailTokenClaim{
		Email:      e.Email,
		AccessType: e.AccessType,
	}
	claim.IssuedAt = now.Unix()
	claim.ExpiresAt = end.Unix()

	signKey := []byte(os.Getenv("JWT_SIGNATURE_KEY"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	tokenString, _ := token.SignedString(signKey)

	return tokenString
}

func VerifyToken(tokenString string) (*models.EmailTokenClaim, error) {
	signKey := []byte(os.Getenv("JWT_SIGNATURE_KEY"))

	claim := new(models.EmailTokenClaim)

	_, err := jwt.ParseWithClaims(tokenString, claim, func(token *jwt.Token) (interface{}, error) {
		return []byte(signKey), nil
	})

	if err != nil {
		return nil, err
	}

	return claim, nil
}
