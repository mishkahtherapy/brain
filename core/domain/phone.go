package domain

import "regexp"

type PhoneNumber string
type WhatsAppNumber string

var whatsAppRegex = regexp.MustCompile(`^(\+[1-9]\d{1,14}|\d{1,15})$`)

func (w WhatsAppNumber) IsValid() bool {
	// WhatsApp number must be 10 digits
	return whatsAppRegex.MatchString(string(w))
}
