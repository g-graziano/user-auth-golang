package http

import (
	"context"
	"net/http"

	"github.com/g-graziano/user-auth-golang/helper"
	"github.com/g-graziano/user-auth-golang/models"
	"github.com/g-graziano/user-auth-golang/service/token"
	json "github.com/json-iterator/go"
)

func HandleVerifyTfa(ctx context.Context, tkn token.Token) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		var otpRequest *models.OTPRequest

		if err := json.NewDecoder(r.Body).Decode(&otpRequest); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.Message(false, "Invalid Request"))
			return
		}

		otpRequest.XID = xid

		err := helper.GetReqHeader(&ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.ErrorMessage(0, err.Error()))
			return
		}

		verified, err := tkn.VerifyTfa(ctx, otpRequest)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.Message(false, "Invalid Request"))
			return
		}

		bs, err := json.ConfigFastest.Marshal(verified)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.Message(false, "Invalid Request"))
			return
		}

		w.Write(bs)
		return
	}
}
