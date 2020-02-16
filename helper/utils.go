package helper

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	qrcode "github.com/skip2/go-qrcode"
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

func GenerateNumericCode(min int, max int) string {
	rand.Seed(time.Now().UnixNano())

	return strconv.Itoa(min + rand.Intn(max-min))
}

func GenerateQRCodeToBase64(value string) (string, error) {
	png, err := qrcode.Encode("value", qrcode.Medium, 256)

	if err != nil {
		return "", err
	}

	base64Text := make([]byte, base64.StdEncoding.EncodedLen(len(png)))
	base64.StdEncoding.Encode(base64Text, []byte(png))

	return string(base64Text), nil
}

func GenerateRandString(n int) string {
	var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func GenerateRandChar(n int) string {
	var letterRunes = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
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
