package domain

type Specialization struct {
	ID        SpecializationID `json:"id"`
	Name      string           `json:"name"`
	CreatedAt UTCTimestamp     `json:"createdAt"`
	UpdatedAt UTCTimestamp     `json:"updatedAt"`
}
