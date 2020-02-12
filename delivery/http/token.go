package http

import (
	"net/http"

	"github.com/g-graziano/userland/helper"
	"github.com/g-graziano/userland/models"
	"github.com/g-graziano/userland/service/token"
	json "github.com/json-iterator/go"
)

func HandleVerifyTfa(tkn token.Token) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get("xid")

		var otpRequest *models.OTPRequest

		if err := json.NewDecoder(r.Body).Decode(&otpRequest); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			helper.Response(w, helper.Message(false, "Invalid Request"))
			return
		}

		otpRequest.XID = xid

		verified, err := tkn.VerifyTfa(otpRequest)
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
