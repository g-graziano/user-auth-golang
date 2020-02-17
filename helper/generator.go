package helper

import (
	"encoding/base64"
	"math/rand"
	"strconv"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

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
