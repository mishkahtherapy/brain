package domain

import (
	"fmt"

	"github.com/google/uuid"
)

func GeneratePrefixedUUID(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, uuid.NewString())
}
