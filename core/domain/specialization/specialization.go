package specialization

import "github.com/mishkahtherapy/brain/core/domain"

type Specialization struct {
	ID        domain.SpecializationID `json:"id"`
	Name      string                  `json:"name"`
	CreatedAt domain.UTCTimestamp     `json:"-"`
	UpdatedAt domain.UTCTimestamp     `json:"-"`
}
