package user

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/g-graziano/user-auth-golang/helper"
	"github.com/g-graziano/user-auth-golang/models"
	"github.com/g-graziano/user-auth-golang/repository/postgres"
	"github.com/g-graziano/user-auth-golang/repository/redis"
	sdg "github.com/g-graziano/user-auth-golang/repository/sendgrid"
	"github.com/go-playground/validator"
	"github.com/rs/xid"
	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
)

type User interface {
	GetAPIClientID(client *models.ClientID) (*models.ClientID, error)
	Login(ctx context.Context, user *models.Login) (*models.AccessToken, error)
	ByPassTfa(ctx context.Context, code *models.OTPRequest) (*models.AccessToken, error)
	GetNewAccessToken(ctx context.Context, token *models.AccessTokenRequest) (*models.AccessToken, error)
	RefreshToken(ctx context.Context, user *models.User) (*models.AccessToken, error)

	Logout(user *models.User) error
	Register(user *models.RegisterRequest) error

	VerifyEmail(user *models.User) error
	SendEmailValidation(user *models.User) error
	ResendEmailValidation(user *models.User) error

	ForgotPassword(user *models.User) error
	ResetPassword(resetPass *models.ResetPass) error

	GetUserProfile(user *models.User) (*models.ProfileResponse, error)
	UpdateUserPicture(picture *models.UploadProfile) error
	GetUserTfaStatus(user *models.User) (*models.TFAStatus, error)
	EnrollTfa(user *models.User) (*models.EnrollTfa, error)
	GetUserEmail(user *models.User) (*models.GetEmailResponse, error)
	UpdateUserProfile(user *models.User) error
	DeleteProfilePicture(user *models.User) error
	UpdateUserPassword(user *models.ChangePassword) error
	RequestChangePassword(user *models.User) error
	DeleteUser(user *models.User) error
	DeleteOtherSession(token *models.AccessTokenRequest) error
	DeleteCurrentSession(token *models.UserToken) error
	RemoveTfa(user *models.User) error
	CheckJWTIsActive(token *models.UserToken) error
	ActivateTfa(secret *models.ActivateTfaRequest) (*models.BackupCodesResponse, error)

	GetListEvent(user *models.User) (*models.ListEventResponse, error)
	GetListSession(session *models.ListSessionRequest) (*models.ListSessionResponse, error)
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

func (u *user) GetListEvent(user *models.User) (*models.ListEventResponse, error) {
	currentUser, err := u.postgres.GetUser(user)

	if err != nil {
		return nil, err
	}

	if len(currentUser) < 1 {
		return nil, errors.New("Invalid auth token")
	}

	limit := 3
	offset := 0

	eventResult, err := u.postgres.GetEvent(
		&models.User{ID: currentUser[0].ID},
		limit,
		offset,
	)

	if err != nil {
		return nil, err
	}

	var listEvent models.ListEventResponse
	var listEventData models.ListEventResponseData

	listEvent.Pagination.Page = 0
	listEvent.Pagination.PerPage = limit

	for _, v := range eventResult {
		listEventData.IP = v.IPAddress
		listEventData.Event = v.Event
		listEventData.UA = v.UA
		listEventData.CreatedAt = v.CreatedAt
		listEventData.Client.ID = v.UserID
		listEventData.Client.Name = v.ClientName

		listEvent.Data = append(listEvent.Data, listEventData)
	}

	return &listEvent, nil
}

func (u *user) GetListSession(session *models.ListSessionRequest) (*models.ListSessionResponse, error) {
	currentUser, err := u.postgres.GetUser(&models.User{XID: session.XID})

	if err != nil {
		return nil, err
	}

	if len(currentUser) < 1 {
		return nil, errors.New("Invalid auth token")
	}

	sessionResult, err := u.postgres.GetSession(&models.UserToken{UserID: currentUser[0].ID})

	if err != nil {
		return nil, err
	}

	var listSession models.ListSessionResponse
	var listSessionData models.ListSessionResponseData

	for _, v := range sessionResult {
		listSessionData.IsCurrent = false

		if session.Token == v.Token {
			listSessionData.IsCurrent = true
		}

		listSessionData.IP = v.IPAddress
		listSessionData.CreatedAt = v.CreatedAt
		listSessionData.UpdatedAt = v.UpdatedAt

		listSessionData.Client.ID = v.UserID
		listSessionData.Client.Name = v.ClientName

		listSession.Data = append(listSession.Data, listSessionData)
	}

	return &listSession, nil
}

func (u *user) CheckJWTIsActive(token *models.UserToken) error {
	currentToken, err := u.postgres.GetToken(token)

	if err != nil {
		return err
	}

	if len(currentToken) < 1 {
		return errors.New("Invalid auth token")
	}

	if currentToken[0].Status == "nonactive" {
		return errors.New("Invalid auth token")
	}

	return nil
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
	OTP := helper.GenerateNumericCode(1000, 9999)

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

func (u *user) GetAPIClientID(client *models.ClientID) (*models.ClientID, error) {
	result, err := u.postgres.GetClientID(client)

	if err != nil {
		return nil, err
	}

	if len(result) < 1 {
		return nil, errors.New("Client API not valid")
	}

	return result[0], nil
}

func (u *user) Login(ctx context.Context, user *models.Login) (*models.AccessToken, error) {
	var loginUser []*models.User

	user.Email = strings.ToLower(user.Email)

	loginUser, err := u.postgres.GetActiveUser(&models.User{Email: user.Email})

	if err != nil {
		return nil, err
	}

	if len(loginUser) < 1 {
		return nil, errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(loginUser[0].Password), []byte(user.Password)); err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return nil, errors.New("invalid login credentials, please try again")
	}

	var token models.TokenClaim

	clientID, err := strconv.ParseUint(fmt.Sprintf("%v", ctx.Value(helper.StringToInterface("client-id"))), 0, 64)
	if err != nil {
		return nil, err
	}

	token.ClientID = clientID
	token.XID = loginUser[0].XID

	if loginUser[0].TFA {
		token.AccessType = "tfa"
		token.ExpiredAt = time.Minute * 5
	} else {
		token.Email = loginUser[0].Email
		token.AccessType = "login"
		token.ExpiredAt = time.Hour * 24
	}

	tokenString := token.TokenGenerator()

	accessToken := &models.AccessToken{
		Value:     tokenString,
		Type:      "Bearer",
		ExpiredAt: time.Now().Add(token.ExpiredAt).String(),
	}

	err = u.postgres.CreateToken(ctx, &models.UserToken{
		Token:        tokenString,
		UserID:       loginUser[0].ID,
		TokenType:    "Bearer",
		RefreshToken: helper.NullStringFunc("", false),
	})

	if err != nil {
		return nil, err
	}

	err = u.postgres.CreateEvent(ctx, "login", loginUser[0].ID)

	if err != nil {
		return nil, err
	}

	if loginUser[0].TFA {
		u.SendEmailOTP(loginUser[0])
	}

	return accessToken, err
}

func (u *user) Logout(user *models.User) error {
	var logoutUser, err = u.postgres.GetUser(&models.User{XID: user.XID})

	if len(logoutUser) < 1 {
		return errors.New("user not found")
	}

	err = u.postgres.UpdateUser(logoutUser[0])
	if err != nil {
		return err
	}
	return nil
}

func (u *user) Register(user *models.RegisterRequest) error {
	if err := user.ValidateRegister(); err != nil {
		return err
	}

	existingUser, err := u.postgres.GetUser(&models.User{Email: user.Email})

	if err != nil {
		return err
	}

	if len(existingUser) > 0 {
		return errors.New("Email Already Exists")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	var createUser models.User

	createUser.Password = string(hashedPassword)
	createUser.Fullname = strings.ToLower(user.Fullname)
	createUser.Email = strings.ToLower(user.Email)
	createUser.XID = xid.New().String()

	err = u.postgres.CreateUser(&createUser)
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

	verifyUser[0].Status = "active"

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

	if len(findUser) < 1 {
		return errors.New("user not found")
	}

	u.SendEmailValidation(findUser[0])

	return nil
}

func (u *user) ForgotPassword(user *models.User) error {
	var forgotUser, err = u.postgres.GetActiveUser(user)

	if err != nil {
		return err
	}

	if len(forgotUser) < 1 {
		return errors.New("user not found")
	}

	var tokenClaim = &models.TokenClaim{
		Email:      forgotUser[0].Email,
		AccessType: "resetpassword",
		ExpiredAt:  time.Minute * 5,
	}

	token := tokenClaim.TokenGenerator()

	var email models.Email
	email.Subject = "Reset Password"
	email.RecipientName = forgotUser[0].Fullname
	email.RecipientEmail = forgotUser[0].Email
	email.PlainContent = "Hi " + forgotUser[0].Fullname + ", Please input token below"
	email.HTMLContent = `<p>Hi ` + forgotUser[0].Fullname + `,</p>
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

func (u *user) RefreshToken(ctx context.Context, user *models.User) (*models.AccessToken, error) {
	var currentUser, err = u.postgres.GetActiveUser(user)

	if err != nil {
		return nil, err
	}

	if len(currentUser) < 1 {
		return nil, errors.New("user not found")
	}

	clientID, err := strconv.ParseUint(fmt.Sprintf("%v", ctx.Value(helper.StringToInterface("client-id"))), 0, 64)
	if err != nil {
		return nil, err
	}

	var tokenClaim = &models.TokenClaim{
		XID:        currentUser[0].XID,
		Email:      currentUser[0].Email,
		AccessType: "refreshtoken",
		ExpiredAt:  time.Hour * 8760,
		ClientID:   clientID,
	}

	tokenString := tokenClaim.TokenGenerator()

	err = u.postgres.CreateToken(ctx, &models.UserToken{Token: tokenString, UserID: currentUser[0].ID, TokenType: "Bearer"})

	if err != nil {
		return nil, err
	}

	err = u.postgres.CreateEvent(ctx, "refresh token", currentUser[0].ID)

	if err != nil {
		return nil, err
	}

	accessToken := &models.AccessToken{
		Value:     tokenString,
		Type:      "Bearer",
		ExpiredAt: time.Now().Add(tokenClaim.ExpiredAt).String(),
	}

	return accessToken, nil
}

func (u *user) GetNewAccessToken(ctx context.Context, token *models.AccessTokenRequest) (*models.AccessToken, error) {
	var currentUser, err = u.postgres.GetActiveUser(&models.User{XID: token.XID})

	if err != nil {
		return nil, err
	}

	if len(currentUser) < 1 {
		return nil, errors.New("user not found")
	}

	clientID, err := strconv.ParseUint(fmt.Sprintf("%v", ctx.Value(helper.StringToInterface("client-id"))), 0, 64)
	if err != nil {
		return nil, err
	}

	var tokenClaim = &models.TokenClaim{
		XID:        currentUser[0].XID,
		Email:      currentUser[0].Email,
		AccessType: "login",
		ExpiredAt:  time.Hour * 24,
		ClientID:   clientID,
	}

	tokenString := tokenClaim.TokenGenerator()

	err = u.postgres.DeleteToken(&models.UserToken{
		Status:       "nonactive",
		RefreshToken: helper.NullStringFunc(token.RefreshToken, true),
	})

	if err != nil {
		return nil, err
	}

	err = u.postgres.CreateToken(ctx, &models.UserToken{
		Token:        tokenString,
		UserID:       currentUser[0].ID,
		TokenType:    "Bearer",
		RefreshToken: helper.NullStringFunc(token.RefreshToken, true),
	})

	if err != nil {
		return nil, err
	}

	err = u.postgres.CreateEvent(ctx, "new access token", currentUser[0].ID)

	if err != nil {
		return nil, err
	}

	accessToken := &models.AccessToken{
		Value:     tokenString,
		Type:      "Bearer",
		ExpiredAt: time.Now().Add(tokenClaim.ExpiredAt).String(),
	}

	return accessToken, nil
}

func (u *user) ByPassTfa(ctx context.Context, codes *models.OTPRequest) (*models.AccessToken, error) {
	var currentUser, err = u.postgres.GetActiveUser(&models.User{XID: codes.XID})

	if err != nil {
		return nil, err
	}

	if len(currentUser) < 1 {
		return nil, errors.New("user not found")
	}

	resultCode, err := u.postgres.GetBackUpCode(&models.BackupCodes{
		UserID: currentUser[0].ID,
		Codes:  codes.Code,
	})

	if err != nil {
		return nil, err
	}

	if len(resultCode) < 1 {
		return nil, errors.New("code not valid")
	}

	clientID, err := strconv.ParseUint(fmt.Sprintf("%v", ctx.Value(helper.StringToInterface("client-id"))), 0, 64)
	if err != nil {
		return nil, err
	}

	var tokenClaim = &models.TokenClaim{
		XID:        currentUser[0].XID,
		Email:      currentUser[0].Email,
		AccessType: "login",
		ExpiredAt:  time.Hour * 24,
		ClientID:   clientID,
	}

	tokenString := tokenClaim.TokenGenerator()

	err = u.postgres.CreateToken(ctx, &models.UserToken{
		Token:     tokenString,
		UserID:    currentUser[0].ID,
		TokenType: "Bearer",
	})

	if err != nil {
		return nil, err
	}

	err = u.postgres.CreateEvent(ctx, "bypass tfa", currentUser[0].ID)

	if err != nil {
		return nil, err
	}

	accessToken := &models.AccessToken{
		Value:     tokenString,
		Type:      "Bearer",
		ExpiredAt: time.Now().Add(tokenClaim.ExpiredAt).String(),
	}

	return accessToken, nil
}

func (u *user) EnrollTfa(user *models.User) (*models.EnrollTfa, error) {
	var currentUser, err = u.postgres.GetActiveUser(&models.User{XID: user.XID})

	if err != nil {
		return nil, err
	}

	if len(currentUser) < 1 {
		return nil, errors.New("user not found")
	}

	if currentUser[0].TFA {
		return nil, errors.New("TFA have already enabled")
	}

	secretCode := helper.GenerateNumericCode(100000, 999999)

	var png []byte
	png, err = qrcode.Encode(secretCode, qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}

	qrString := "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
	secret := helper.GenerateRandString(20)

	err = u.redis.Create(&models.OTP{Key: user.XID + "-secret", Value: secret, Expire: time.Now().Add(time.Minute * 60)})
	if err != nil {
		return nil, err
	}

	err = u.redis.Create(&models.OTP{Key: user.XID + "-code", Value: secretCode, Expire: time.Now().Add(time.Minute * 60)})
	if err != nil {
		return nil, err
	}

	return &models.EnrollTfa{Secret: secret, Qr: qrString}, nil
}

func (u *user) ActivateTfa(secret *models.ActivateTfaRequest) (*models.BackupCodesResponse, error) {
	var currentUser, err = u.postgres.GetActiveUser(&models.User{XID: secret.XID})

	if err != nil {
		return nil, err
	}

	if len(currentUser) < 1 {
		return nil, errors.New("user not found")
	}

	secretID, err := u.redis.Get(&models.OTP{Key: secret.XID + "-secret"})

	if err != nil {
		return nil, err
	}

	if secretID != secret.Secret {
		return nil, errors.New("secret or code not valid")
	}

	code, err := u.redis.Get(&models.OTP{Key: secret.XID + "-code"})

	if err != nil {
		return nil, err
	}

	if code != secret.Code {
		return nil, errors.New("secret or code not valid")
	}

	var backupcodes models.BackupCodesResponse

	for i := 0; i < 2; i++ {
		BUCodes := helper.GenerateRandChar(12)
		err = u.postgres.CreateBackUpCode(&models.BackupCodes{
			UserID: currentUser[0].ID,
			Codes:  BUCodes,
		})

		if err != nil {
			return nil, err
		}

		backupcodes.BackupCodes = append(backupcodes.BackupCodes, BUCodes)
	}

	currentUser[0].TFA = true

	err = u.postgres.UpdateUser(currentUser[0])
	if err != nil {
		return nil, err
	}

	return &backupcodes, nil
}

func (u *user) DeleteOtherSession(token *models.AccessTokenRequest) error {
	var currentUser, err = u.postgres.GetActiveUser(&models.User{XID: token.XID})

	if err != nil {
		return err
	}

	if len(currentUser) < 1 {
		return errors.New("user not found")
	}

	err = u.postgres.DeleteToken(&models.UserToken{
		UserID: currentUser[0].ID,
		Status: "nonactive",
		Token:  token.RefreshToken,
	})

	if err != nil {
		return err
	}

	return nil
}

func (u *user) DeleteCurrentSession(token *models.UserToken) error {
	err := u.postgres.DeleteToken(&models.UserToken{
		Status: "nonactive",
		Token:  token.Token,
	})

	if err != nil {
		return err
	}

	return nil
}

func (u *user) RequestChangePassword(user *models.User) error {
	var getUser, err = u.postgres.GetActiveUser(&models.User{XID: user.XID})

	if err != nil {
		return err
	}

	if len(getUser) < 1 {
		return errors.New("user not found")
	}

	var tokenClaim = &models.TokenClaim{
		XID:        getUser[0].XID,
		Email:      user.Email,
		AccessType: "refreshtoken",
		ExpiredAt:  time.Minute * 10,
	}

	token := tokenClaim.TokenGenerator()

	var email models.Email
	email.Subject = "Change Email"
	email.RecipientName = getUser[0].Fullname
	email.RecipientEmail = user.Email
	email.PlainContent = "Hi " + getUser[0].Fullname + ", Please click link below to verify your email address so we know that it's really you!"
	email.HTMLContent = `<p>Hi ` + user.Fullname + `,</p>
		<p>Please click link below to verify your email address so we know that it's really you!</p>
		<p><a href="http://0.0.0.0:8080/me/change-email/` + token + `" style="box-sizing: border-box;
		border-color: #ED3237;font-weight: 400;text-decoration: none;display: inline-block;margin: 0;color: #ffffff;background-color: #ED3237;
		border: solid 1px #ED3237;border-radius: 2px;font-size: 14px;padding: 12px 45px;">Confirm Email Address<a></p>`

	err = sdg.SendEmail(&email)
	if err != nil {
		return err
	}

	return nil
}

func (u *user) ResetPassword(resetPass *models.ResetPass) error {
	token, err := models.VerifyToken(resetPass.Token)

	if err != nil {
		return err
	}

	validate := validator.New()

	if err := validate.VarWithValue(resetPass.Password, resetPass.PasswordConfirm, "eqfield"); err != nil {
		return errors.New("Password and password confirm must same")
	}

	if token.AccessType != "resetpassword" {
		return errors.New("token tidak valid")
	}

	resetUser, err := u.postgres.GetActiveUser(&models.User{Email: token.Email})

	if err != nil {
		return err
	}

	if len(resetUser) < 1 {
		return errors.New("user not found")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(resetPass.Password), bcrypt.DefaultCost)

	resetUser[0].Password = string(hashedPassword)

	err = u.postgres.UpdateUser(resetUser[0])
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

func (u *user) GetUserTfaStatus(user *models.User) (*models.TFAStatus, error) {
	var foundUser, err = u.postgres.GetUser(user)

	if err != nil {
		return nil, err
	}

	if len(foundUser) < 1 {
		return nil, errors.New("user not found")
	}

	return &models.TFAStatus{
		Enabled:   foundUser[0].TFA,
		EnabledAt: foundUser[0].EnabledTfaAt.Time,
	}, nil
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

func (u *user) UpdateUserPicture(picture *models.UploadProfile) error {
	var foundUser, err = u.postgres.GetUser(&models.User{XID: picture.UserXID})

	if err != nil {
		return err
	}

	if len(foundUser) < 1 {
		return errors.New("user not found")
	}

	var buf = new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	part, _ := writer.CreateFormFile("image", "test")
	io.Copy(part, picture.File)

	writer.Close()
	req, _ := http.NewRequest("POST", "https://api.imgur.com/3/upload", buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Client-ID ff15fec03c2be0e")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)

	var response models.ImgurUploadResponse

	err = json.Unmarshal(b, &response)
	if err != nil {
		return err
	}

	foundUser[0].Picture = helper.NullStringFunc(response.Data.Link, true)

	err = u.postgres.UpdateUser(foundUser[0])
	if err != nil {
		return err
	}

	return nil
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

func (u *user) DeleteProfilePicture(user *models.User) error {
	var foundUser, err = u.postgres.GetUser(user)

	if err != nil {
		return err
	}

	if len(foundUser) < 1 {
		return errors.New("user not found")
	}

	foundUser[0].Picture = helper.NullStringFunc("", false)

	err = u.postgres.UpdateUser(foundUser[0])
	if err != nil {
		return err
	}

	return nil
}

func (u *user) DeleteUser(user *models.User) error {
	var foundUser, err = u.postgres.GetUser(user)

	if err != nil {
		return err
	}

	if len(foundUser) < 1 {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser[0].Password), []byte(user.Password)); err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return errors.New("password not valid")
	}

	foundUser[0].Status = "deleted"

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
		return errors.New("Password and password confirm must same")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	foundUser[0].Password = string(hashedPassword)

	err = u.postgres.UpdateUser(foundUser[0])
	if err != nil {
		return err
	}

	return nil
}

func (u *user) RemoveTfa(user *models.User) error {
	var foundUser, err = u.postgres.GetUser(user)

	if err != nil {
		return err
	}

	if len(foundUser) < 1 {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser[0].Password), []byte(user.Password)); err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return errors.New("current password not valid")
	}

	foundUser[0].TFA = false

	err = u.postgres.UpdateUser(foundUser[0])
	if err != nil {
		return err
	}

	return nil
}
