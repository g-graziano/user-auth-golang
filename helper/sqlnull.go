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

// NullBool is an alias for sql.NullBool data type
type NullBool struct{ sql.NullBool }

var _ nullable = (*NullBool)(nil)

// MarshalJSON for NullBool
func (nb *NullBool) MarshalJSON() ([]byte, error) { return fromSQL(nb.Bool, nb.Valid) }

// UnmarshalJSON for NullBool
func (nb *NullBool) UnmarshalJSON(b []byte) error { return toSQL(b, &nb.Bool, &nb.Valid) }

// NullFloat64 is an alias for sql.NullFloat64 data type
type NullFloat64 struct{ sql.NullFloat64 }

var _ nullable = (*NullFloat64)(nil)

// MarshalJSON for NullFloat64
func (nf *NullFloat64) MarshalJSON() ([]byte, error) { return fromSQL(nf.Float64, nf.Valid) }

// UnmarshalJSON for NullFloat64
func (nf *NullFloat64) UnmarshalJSON(b []byte) error { return toSQL(b, &nf.Float64, &nf.Valid) }

// NullInt32 is an alias for sql.NullInt32 data type
type NullInt32 struct{ sql.NullInt32 }

var _ nullable = (*NullInt32)(nil)

// MarshalJSON for NullInt32
func (ni *NullInt32) MarshalJSON() ([]byte, error) { return fromSQL(ni.Int32, ni.Valid) }

// UnmarshalJSON for NullInt32
func (ni *NullInt32) UnmarshalJSON(b []byte) error { return toSQL(b, &ni.Int32, &ni.Valid) }

// NullInt64 is an alias for sql.NullInt64 data type
type NullInt64 struct{ sql.NullInt64 }

var _ nullable = (*NullInt64)(nil)

// MarshalJSON for NullInt64
func (ni *NullInt64) MarshalJSON() ([]byte, error) { return fromSQL(ni.Int64, ni.Valid) }

// UnmarshalJSON for NullInt64
func (ni *NullInt64) UnmarshalJSON(b []byte) error { return toSQL(b, &ni.Int64, &ni.Valid) }

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
