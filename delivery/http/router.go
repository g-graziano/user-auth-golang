package http

import (
	"net/http"
	"os"
	"time"

	"github.com/g-graziano/userland/middleware"
	"github.com/g-graziano/userland/service/token"
	"github.com/g-graziano/userland/service/user"
	"github.com/go-chi/chi"
	mdlw "github.com/go-chi/chi/middleware"
)

func Router(user user.User, token token.Token) {
	r := chi.NewRouter()
	r.Use(mdlw.Logger)
	r.Use(mdlw.Timeout(60 * time.Second))
	r.Use(mdlw.Recoverer)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", HandleUserRegister(user))
		r.Get("/verification/{xid}", HandleEmailVerification(user))
		r.Post("/verification", HandleRequestEmailVerification(user))
		r.Post("/login", HandleLogin(user))
		r.Post("/password/forgot", HandleForgotPassword(user))
		r.Post("/password/reset", HandleResetPassword(user))
		r.With(middleware.JwtAuthentication).Group(func(r chi.Router) {
			r.Post("/tfa/verify", HandleVerifyTfa(token))
		})
	})

	r.Route("/me", func(r chi.Router) {
		r.With(middleware.JwtTfaAuthentication).Group(func(r chi.Router) {
			r.Get("/", HandleGetUserProfile(user))
			r.Post("/", HandleUpdateUserProfile(user))
			r.Get("/email", HandleGetUserEmail(user))
			r.Post("/password", HandleUpdatePassword(user))
			r.Post("/picture", HandleSetProfilePicture(user))
		})
	})

	http.ListenAndServe(os.Getenv("SERVER_ADDRESS"), r)

	// r := mux.NewRouter()

	// authH := r.PathPrefix("/auth").Subrouter()
	// authH.HandleFunc("/register", HandleUserRegister(user))
	// authH.HandleFunc("/login", HandleLogin(user))
	// authH.HandleFunc("/verification", HandleEmailVerification(user)).Methods("GET")
	// authH.HandleFunc("/verification", HandleRequestEmailVerification(user)).Methods("POST")

	// authH.HandleFunc("/tfa/verify", HandleUserRegister(user))
	// authH.HandleFunc("/tfa/bypass", HandleUserRegister(user))

	// authH.HandleFunc("/password/forgot", HandleUserRegister(user))
	// authH.HandleFunc("/password/reset", HandleUserRegister(user))

	// // Me
	// meH := r.PathPrefix("/me").Subrouter()
	// meH.Handle("", middleware.JwtAuthentication(HandleGetUserProfile(user))).Methods("GET")
	// // meH.HandleFunc("", HandleUpdateUserProfile(user)).Methods("POST")

	// meH.HandleFunc("/email", HandleUserRegister(user)).Methods("GET")
	// meH.HandleFunc("/email", HandleUserRegister(user)).Methods("POST")

	// meH.HandleFunc("/password", HandleUserRegister(user)).Methods("POST")

	// meH.HandleFunc("/picture", HandleUserRegister(user)).Methods("POST")
	// meH.HandleFunc("/picture", HandleUserRegister(user)).Methods("DELETE")

	// meH.HandleFunc("/tfa", HandleUserRegister(user)).Methods("GET")
	// meH.HandleFunc("/tfa/enroll", HandleUserRegister(user)).Methods("GET")
	// meH.HandleFunc("/tfa/enroll", HandleUserRegister(user)).Methods("POST")
	// meH.HandleFunc("/tfa/remove", HandleUserRegister(user)).Methods("GET")

	// meH.HandleFunc("/events", HandleUserRegister(user)).Methods("GET")
	// meH.HandleFunc("/delete", HandleUserRegister(user)).Methods("POST")

	// meH.HandleFunc("/session", HandleUserRegister(user)).Methods("GET")
	// meH.HandleFunc("/session", HandleUserRegister(user)).Methods("DELETE")
	// meH.HandleFunc("/session/other", HandleUserRegister(user)).Methods("DELETE")
	// meH.HandleFunc("/session/refresh_token", HandleUserRegister(user)).Methods("GET")
	// meH.HandleFunc("/session/access_token", HandleUserRegister(user)).Methods("GET")

	// // userH := r.PathPrefix("/user").Subrouter()
	// // userH.HandleFunc("/verification/{xid}", HandleEmailVerification(user))
	// // userH.Handle("/search-email/{email}", middleware.JwtAuthentication(HandleSearchUserByEmail(user))).Methods("GET")
	// // userH.Handle("/profile", middleware.JwtAuthentication(HandleGetUserProfile(user))).Methods("GET")
	// // userH.Handle("/logout", middleware.JwtAuthentication(HandleLogout(user)))

	// headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	// originsOk := handlers.AllowedOrigins([]string{"*"})
	// methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"})

	// http.ListenAndServe(os.Getenv("SERVER_ADDRESS"), handlers.CORS(headersOk, originsOk, methodsOk)(r))
}
