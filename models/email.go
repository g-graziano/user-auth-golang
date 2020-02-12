package models

type Email struct {
	RecipientName  string
	RecipientEmail string
	Subject        string
	HTMLContent    string
	PlainContent   string
}
