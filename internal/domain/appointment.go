package domain

import (
	"time"

	"github.com/google/uuid"
)

type AppointmentStatus string

const (
	StatusPending   AppointmentStatus = "PENDING"
	StatusConfirmed AppointmentStatus = "CONFIRMED"
	StatusCancelled AppointmentStatus = "CANCELLED"
	StatusCompleted AppointmentStatus = "COMPLETED"
)

type Appointment struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	// Foreign Keys
	CustomerID uuid.UUID `gorm:"type:uuid;not null;index"`
	Customer   User      `gorm:"foreignKey:CustomerID"` // Relation

	ProviderID uuid.UUID `gorm:"type:uuid;not null;index"`
	Provider   User      `gorm:"foreignKey:ProviderID"` // Relation

	// Service Details
	ServiceType string    `gorm:"type:varchar(50);not null"` // e.g., "Haircut", "Consulting"
	StartTime   time.Time `gorm:"not null;index"`
	EndTime     time.Time `gorm:"not null"`

	Status AppointmentStatus `gorm:"type:varchar(20);default:'PENDING'"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
