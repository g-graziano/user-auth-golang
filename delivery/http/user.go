package http

import (
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
			helper.Response(w, helper.Message(false, err.Error()))
			return
		}

		err := user.Register(newUser)
		if err != nil {
			w.WriteHeader(http.StatusAccepted)
			helper.Response(w, helper.Message(false, err.Error()))

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
			helper.Response(w, helper.Message(false, err.Error()))

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
			helper.Response(w, helper.Message(false, err.Error()))
			return
		}

		if input.Type == "email" {
			err := user.ResendEmailValidation(&models.User{Email: input.Recipient})

			if err != nil {
				w.WriteHeader(http.StatusAccepted)
				helper.Response(w, helper.Message(false, err.Error()))

				return
			}

		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Request Email Verification Success!"))

		return
	}
}

func HandleLogin(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginUser *models.User
		if err := json.NewDecoder(r.Body).Decode(&loginUser); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.Message(false, err.Error()))
			return
		}

		login, err := user.Login(loginUser)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.Message(false, err.Error()))
			return
		}

		bs, err := json.ConfigFastest.Marshal(login)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.Message(false, err.Error()))
			return
		}

		w.Write(bs)
	}
}

func HandleLogout(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")
		token := r.Header.Get("token")

		err := user.Logout(&models.User{XID: xid, Token: helper.NullStringFunc(token, true)})
		if err != nil {
			panic(err)
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
			helper.Response(w, helper.Message(false, err.Error()))
			return
		}

		err := user.ForgotPassword(forgotUser)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.Message(false, err.Error()))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Request Forgot Password Success!"))

		return
	}
}

func HandleResetPassword(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resetPass *models.ResetPass
		if err := json.NewDecoder(r.Body).Decode(&resetPass); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.Message(false, err.Error()))
			return
		}

		err := user.ResetPassword(resetPass)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.Message(false, err.Error()))
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
			helper.Response(w, helper.Message(false, err.Error()))
			return
		}

		bs, err := json.ConfigFastest.Marshal(profile)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.Message(false, err.Error()))
			return
		}

		w.Write(bs)

		return
	}
}

func HandleGetUserEmail(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		email, err := user.GetUserEmail(&models.User{XID: xid})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.Message(false, err.Error()))
			return
		}

		bs, err := json.ConfigFastest.Marshal(email)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.Message(false, err.Error()))
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
			helper.Response(w, helper.Message(false, err.Error()))
			return
		}

		update.XID = xid

		err := user.UpdateUserProfile(update)
		if err != nil {
			w.WriteHeader(http.StatusAccepted)
			helper.Response(w, helper.Message(false, "Update failed!"))

			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Update Success!"))

		return
	}
}

func HandleUpdatePassword(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		var update *models.ChangePassword
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.Message(false, err.Error()))
			return
		}

		update.XID = xid

		err := user.UpdateUserPassword(update)
		if err != nil {
			w.WriteHeader(http.StatusAccepted)
			helper.Response(w, helper.Message(false, err.Error()))

			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Update Success!"))

		return
	}
}

func HandleSetProfilePicture(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		var update *models.ChangePassword
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.Message(false, err.Error()))
			return
		}

		update.XID = xid

		err := user.UpdateUserPassword(update)
		if err != nil {
			w.WriteHeader(http.StatusAccepted)
			helper.Response(w, helper.Message(false, err.Error()))

			return
		}

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Update Success!"))

		return
	}
}
