package booking

import "errors"

var (
	ErrBookingAlreadyConfirmed = errors.New("booking is already confirmed")
	ErrFailedToCreateSession   = errors.New("failed to create session for confirmed booking")
)
