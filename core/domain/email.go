package domain

type Email string

func NewEmail(email string) Email {
	return Email(email)
}

func (e Email) String() string {
	return string(e)
}
