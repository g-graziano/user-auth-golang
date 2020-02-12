package helper

import (
	"encoding/json"
	"net/http"
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
