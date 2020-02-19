package http

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/g-graziano/user-auth-golang/middleware"
	"github.com/g-graziano/user-auth-golang/service/token"
	"github.com/g-graziano/user-auth-golang/service/user"
	"github.com/go-chi/chi"
	mdlw "github.com/go-chi/chi/middleware"
)

func Router(ctx context.Context, user user.User, token token.Token) {
	r := chi.NewRouter()
	r.Use(mdlw.Logger)
	r.Use(mdlw.Timeout(60 * time.Second))
	r.Use(mdlw.Recoverer)

	r.Route("/auth", func(r chi.Router) {
		r.With(middleware.APIClientAuthentication(user)).Group(func(r chi.Router) {
			r.Post("/register", HandleUserRegister(user))
			r.Post("/login", HandleLogin(ctx, user))

			r.Post("/verification", HandleRequestEmailVerification(user))

			r.Post("/password/forgot", HandleForgotPassword(user))
			r.Post("/password/reset", HandleResetPassword(user))
		})

		r.Get("/verification/{xid}", HandleEmailVerification(user))

		r.With(middleware.JwtTfaAuthentication).Group(func(r chi.Router) {
			r.Post("/tfa/bypass", HandleByPassTfa(ctx, user))
			r.Post("/tfa/verify", HandleVerifyTfa(ctx, token))
		})
	})

	r.Route("/me", func(r chi.Router) {
		r.With(middleware.JwtAuthentication(user)).Group(func(r chi.Router) {
			r.Get("/", HandleGetUserProfile(user))
			r.Post("/", HandleUpdateUserProfile(user))

			r.Get("/email", HandleGetUserEmail(user))
			r.Post("/email", HandleRequestChangePassword(user))

			r.Post("/password", HandleUpdatePassword(user))

			r.Delete("/picture", HandleDeleteProfilePicture(user))
			r.Post("/picture", HandleSetProfilePicture(user))

			r.Post("/delete", HandleDeleteUser(user))

			r.Get("/tfa", HandleGetTfaStatus(user))
			r.Post("/tfa/remove", HandleRemoveTfa(user))
			r.Get("/tfa/enroll", HandleTfaEnroll(user))
			r.Post("/tfa/enroll", HandleActivateTfa(user))

			r.Get("/session", HandleGetListSession(user))
			r.Get("/session/refresh_token", HandleGetRefreshToken(ctx, user))
			r.Delete("/session/other", HandleDeleteOtherSession(user))
			r.Delete("/session", HandleEndCurrentSession(user))

			r.Get("/events", HandleGetListEvent(user))
		})

		r.With(middleware.JwtACTAuthentication).Get("/session/access_token", HandleGetNewAccessToken(ctx, user))
	})

	http.ListenAndServe(os.Getenv("SERVER_ADDRESS"), r)
}
