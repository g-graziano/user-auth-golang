package helper

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

func Message(status bool, message string) map[string]interface{} {
	return map[string]interface{}{"success": status, "message": message}
}

func ErrorMessage(code int, message string) map[string]interface{} {
	return map[string]interface{}{"code": code, "message": message}
}

func Response(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(data)
}

func StringToInterface(str string) interface{} {
	return str
}

func GetReqHeader(ctx *context.Context, r *http.Request) error {
	ipAddress := r.Header.Get("X-FORWARDED-FOR")
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}

	clientID, err := strconv.ParseUint(r.Header.Get("client-id"), 10, 64)
	if err != nil {
		return err
	}

	*ctx = context.WithValue(*ctx, StringToInterface("ip-address"), ipAddress)
	*ctx = context.WithValue(*ctx, StringToInterface("user-agent"), r.Header.Get("User-Agent"))
	*ctx = context.WithValue(*ctx, StringToInterface("client-id"), clientID)

	return nil
}
