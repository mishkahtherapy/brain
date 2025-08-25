package config

import "github.com/mishkahtherapy/brain/core/domain"

const minimumBookingTime = domain.DurationMinutes(15)

type BookingConfig struct{}

func GetBookingConfig() BookingConfig {
	return BookingConfig{}
}

func (c *BookingConfig) MinimumBookingTime() domain.DurationMinutes {
	return minimumBookingTime
}
