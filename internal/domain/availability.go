package domain

import (
	"time"

	"github.com/google/uuid"
)

type Availability struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ProviderID uuid.UUID `gorm:"type:uuid;not null;index"`
	DayOfWeek  int       `gorm:"not null"`
	StartTime  string    `gorm:"type:varchar(5);not null"`
	EndTime    string    `gorm:"type:varchar(5);not null"`
	CreatedAt  time.Time
}
