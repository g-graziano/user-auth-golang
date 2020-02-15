package helper

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var null = []byte(`null`)
var ok error = nil
var timeFormat = time.RFC3339

type nullable interface {
	json.Marshaler
	json.Unmarshaler
}

func fromSQL(v interface{}, valid bool) ([]byte, error) {
	if !valid {
		return null, ok
	}
	return jsoniter.ConfigFastest.Marshal(v)
}

func toSQL(b []byte, v interface{}, valid *bool) error {
	var err = jsoniter.ConfigFastest.Unmarshal(b, v)
	*valid = (err == nil)
	return err
}

// NullString is an alias for sql.NullString data type
type NullString struct{ sql.NullString }

func NullStringFunc(s string, b bool) NullString {
	return NullString{sql.NullString{String: s, Valid: b}}
}

var _ nullable = (*NullString)(nil)

// MarshalJSON for NullString
func (ns *NullString) MarshalJSON() ([]byte, error) { return fromSQL(ns.String, ns.Valid) }

// UnmarshalJSON for NullString
func (ns *NullString) UnmarshalJSON(b []byte) error { return toSQL(b, &ns.String, &ns.Valid) }

// NullTime is an alias for sql.NullTime data type
type NullTime struct{ sql.NullTime }

var _ nullable = (*NullTime)(nil)

// MarshalJSON for NullTime
func (nt *NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return null, ok
	}
	return []byte(fmt.Sprintf("\"%s\"", nt.Time.Format(timeFormat))), nil
}

// UnmarshalJSON for NullTime
func (nt *NullTime) UnmarshalJSON(b []byte) error {
	var x, err = time.Parse(timeFormat, string(b))
	if err != nil {
		nt.Valid = false
		return err
	}

	nt.Time = x
	nt.Valid = true
	return nil
}
