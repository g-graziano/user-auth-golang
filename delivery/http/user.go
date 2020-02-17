package http

import (
	"bytes"
	"context"
	"image"
	"io"
	"net/http"

	"github.com/g-graziano/userland/helper"
	"github.com/g-graziano/userland/models"
	"github.com/g-graziano/userland/service/user"
	"github.com/go-chi/chi"
	json "github.com/json-iterator/go"
)

func HandleUserRegister(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newUser *models.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		err := user.Register(newUser)
		if err != nil {
			w.WriteHeader(http.StatusAccepted)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))

			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Register Success!"))

		return
	}
}

func HandleEmailVerification(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := chi.URLParam(r, "xid")

		err := user.VerifyEmail(&models.User{XID: xid})
		if err != nil {
			w.WriteHeader(http.StatusAccepted)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))

			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Email Verification Success!"))

		return
	}
}

func HandleRequestEmailVerification(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input *models.VerificationRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		if input.Type == "email" {
			err := user.ResendEmailValidation(&models.User{Email: input.Recipient})

			if err != nil {
				w.WriteHeader(http.StatusAccepted)
				helper.Response(w, helper.ErrorMessage(0, err.Error()))

				return
			}

		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Request Email Verification Success!"))

		return
	}
}

func HandleLogin(ctx context.Context, user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginUser *models.Login
		if err := json.NewDecoder(r.Body).Decode(&loginUser); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		err := helper.GetReqHeader(&ctx, r)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		login, err := user.Login(ctx, loginUser)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		bs, err := json.ConfigFastest.Marshal(login)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		w.Write(bs)
	}
}

func HandleLogout(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		err := user.Logout(&models.User{XID: xid})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Logout Success"))

		return
	}
}

func HandleForgotPassword(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var forgotUser *models.User
		if err := json.NewDecoder(r.Body).Decode(&forgotUser); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		err := user.ForgotPassword(forgotUser)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Request Forgot Password Success!"))

		return
	}
}

func HandleRequestChangePassword(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		var changeUser *models.User
		if err := json.NewDecoder(r.Body).Decode(&changeUser); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		changeUser.XID = xid

		err := user.RequestChangePassword(changeUser)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Request Email Change Success!"))

		return
	}
}

func HandleResetPassword(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resetPass *models.ResetPass
		if err := json.NewDecoder(r.Body).Decode(&resetPass); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		err := user.ResetPassword(resetPass)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Reset Password Success!"))

		return
	}
}

func HandleGetUserProfile(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		profile, err := user.GetUserProfile(&models.User{XID: xid})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		bs, err := json.ConfigFastest.Marshal(profile)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		w.Write(bs)

		return
	}
}

func HandleGetTfaStatus(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		profile, err := user.GetUserTfaStatus(&models.User{XID: xid})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		bs, err := json.ConfigFastest.Marshal(profile)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		w.Write(bs)

		return
	}
}

func HandleGetRefreshToken(ctx context.Context, user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		err := helper.GetReqHeader(&ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		token, err := user.RefreshToken(ctx, &models.User{XID: xid})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		bs, err := json.ConfigFastest.Marshal(token)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		w.Write(bs)

		return
	}
}

func HandleGetNewAccessToken(ctx context.Context, user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")
		refreshToken := r.Header.Get("token")

		err := helper.GetReqHeader(&ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		token, err := user.GetNewAccessToken(ctx, &models.AccessTokenRequest{XID: xid, RefreshToken: refreshToken})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		bs, err := json.ConfigFastest.Marshal(token)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		w.Write(bs)

		return
	}
}

func HandleDeleteOtherSession(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")
		currentToken := r.Header.Get("token")

		err := user.DeleteOtherSession(&models.AccessTokenRequest{XID: xid, RefreshToken: currentToken})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Other session deleted!"))

		return
	}
}

func HandleEndCurrentSession(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentToken := r.Header.Get("token")

		err := user.DeleteCurrentSession(&models.UserToken{Token: currentToken})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Session ended!"))

		return
	}
}

func HandleGetUserEmail(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		email, err := user.GetUserEmail(&models.User{XID: xid})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		bs, err := json.ConfigFastest.Marshal(email)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		w.Write(bs)

		return
	}
}

func HandleGetListEvent(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		listEvent, err := user.GetListEvent(&models.User{XID: xid})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		bs, err := json.ConfigFastest.Marshal(listEvent)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		w.Write(bs)

		return
	}
}

func HandleUpdateUserProfile(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		var update *models.User
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		update.XID = xid

		err := user.UpdateUserProfile(update)
		if err != nil {
			w.WriteHeader(http.StatusAccepted)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))

			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Update Success!"))

		return
	}
}

func HandleActivateTfa(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		var secret *models.ActivateTfaRequest
		if err := json.NewDecoder(r.Body).Decode(&secret); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		secret.XID = xid

		codes, err := user.ActivateTfa(secret)
		if err != nil {
			w.WriteHeader(http.StatusAccepted)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))

			return
		}

		bs, err := json.ConfigFastest.Marshal(codes)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		w.Write(bs)

		return
	}
}

func HandleUpdatePassword(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		var update *models.ChangePassword
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		update.XID = xid

		err := user.UpdateUserPassword(update)
		if err != nil {
			w.WriteHeader(http.StatusAccepted)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))

			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Update success!"))

		return
	}
}

func HandleTfaEnroll(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		var currentUser models.User
		currentUser.XID = xid

		secretCode, err := user.EnrollTfa(&currentUser)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))

			return
		}

		bs, err := json.ConfigFastest.Marshal(secretCode)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		w.Write(bs)

		return
	}
}

func HandleSetProfilePicture(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 200000) //200 Kb

		if err := r.ParseMultipartForm(200000); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))

			return
		}

		xid := r.Header.Get("xid")

		data := &models.UploadProfile{}
		f, h, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}
		defer f.Close()

		// get image dimension
		img, _, _ := image.DecodeConfig(f)

		defer f.Close()
		fopen, _ := h.Open()
		defer fopen.Close()
		content := bytes.NewBuffer(nil)
		if _, err := io.Copy(content, fopen); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		if http.DetectContentType(content.Bytes()) != "image/jpeg" {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, "image must jpeg type"))
			return
		}

		data.File = content
		data.ContentType = http.DetectContentType(content.Bytes())
		data.Size = int64(content.Len())
		data.Width = int64(img.Width)
		data.Height = int64(img.Height)
		data.UserXID = xid

		err = user.UpdateUserPicture(data)
		if err != nil {
			w.WriteHeader(http.StatusAccepted)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))

			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Success!"))

		return
	}
}

func HandleDeleteProfilePicture(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		err := user.DeleteProfilePicture(&models.User{XID: xid})
		if err != nil {
			w.WriteHeader(http.StatusAccepted)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))

			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Delete picture success!"))

		return
	}
}

func HandleDeleteUser(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		var delete *models.User
		if err := json.NewDecoder(r.Body).Decode(&delete); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		delete.XID = xid

		err := user.DeleteUser(delete)
		if err != nil {
			w.WriteHeader(http.StatusAccepted)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))

			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Delete account success!"))

		return
	}
}

func HandleRemoveTfa(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		var currentUser *models.User
		if err := json.NewDecoder(r.Body).Decode(&currentUser); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		currentUser.XID = xid

		err := user.RemoveTfa(currentUser)
		if err != nil {
			w.WriteHeader(http.StatusAccepted)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))

			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Tfa disabled!"))

		return
	}
}

func HandleByPassTfa(ctx context.Context, user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		var currentUser *models.OTPRequest

		if err := json.NewDecoder(r.Body).Decode(&currentUser); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		currentUser.XID = xid

		err := helper.GetReqHeader(&ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		byPass, err := user.ByPassTfa(ctx, currentUser)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		bs, err := json.ConfigFastest.Marshal(byPass)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, "Invalid Request"))
			return
		}

		w.Write(bs)
		return
	}
}

func HandleGetListSession(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")
		token := r.Header.Get("token")

		listSessionRequest := &models.ListSessionRequest{Token: token, XID: xid}

		listSession, err := user.GetListSession(listSessionRequest)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		bs, err := json.ConfigFastest.Marshal(listSession)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, "Invalid Request"))
			return
		}

		w.Write(bs)
		return
	}
}
