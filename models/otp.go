package models

import "time"

type OTP struct {
	Key    string    `json:"key"`
	Value  string    `json:"value"`
	Expire time.Time `json:"expire"`
}

type OTPRequest struct {
	Code string
	XID  string
}
