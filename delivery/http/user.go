package http

import (
	"net/http"

	"github.com/g-graziano/userland/helper"
	"github.com/g-graziano/userland/models"
	"github.com/g-graziano/userland/service/user"
	"github.com/gorilla/mux"
	json "github.com/json-iterator/go"
)

func HandleUserRegister(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newUser *models.User
		if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.Message(false, "Invalid Request"))
			return
		}

		registeredUser, err := user.Register(newUser)
		if err != nil {
			panic(err)
		}

		res, err := json.ConfigFastest.Marshal(registeredUser)
		if err != nil {
			panic(err)
		}

		w.Write(res)
	}
}

func HandleEmailVerification(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		user.VerifyEmail(&models.User{XID: params["xid"]})

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Email Verification Success!"))
	}
}

func HandleRequestEmailVerification(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		user.SendEmailValidation(&models.User{XID: params["xid"]})

		w.WriteHeader(http.StatusAccepted)
		helper.Response(w, helper.Message(true, "Request Email Verification Success!"))
	}
}

func HandleLogin(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginUser *models.User
		if err := json.NewDecoder(r.Body).Decode(&loginUser); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.Message(false, "Invalid Request"))
			return
		}

		login, err := user.Login(loginUser)
		if err != nil {
			panic(err)
		}

		bs, err := json.ConfigFastest.Marshal(login)
		if err != nil {
			panic(err)
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

func HandleSearchUserByEmail(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		xid := r.Header.Get("xid")

		foundUser, err := user.SearchUserByEmail(xid, &models.User{Email: params["email"]})
		if err != nil {
			panic(err)
		}

		bs, err := json.ConfigFastest.Marshal(foundUser)
		if err != nil {
			panic(err)
		}

		w.Write(bs)

		return
	}
}

func HandleGetUserProfile(user user.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		profile, err := user.GetUserProfile(&models.User{XID: xid})
		if err != nil {
			panic(err)
		}

		bs, err := json.ConfigFastest.Marshal(profile)
		if err != nil {
			panic(err)
		}

		w.Write(bs)

		return
	}
}
