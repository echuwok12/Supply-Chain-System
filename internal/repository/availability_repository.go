package repository

import (
	"appointment-booking/internal/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AvailabilityRepository struct {
	db *gorm.DB
}

func NewAvailabilityRepository(db *gorm.DB) *AvailabilityRepository {
	return &AvailabilityRepository{db: db}
}

func (r *AvailabilityRepository) Save(availability *domain.Availability) error {
	return r.db.Create(availability).Error
}

func (r *AvailabilityRepository) GetByProviderAndDay(providerID uuid.UUID, day int) (*domain.Availability, error) {
	var availability domain.Availability
	err := r.db.Where("provider_id = ? AND day_of_week = ?", providerID, day).First(&availability).Error
	return &availability, err
}

// Return all confirmed/pending appointments for a specific date
func (r *AppointmentRepository) GetProviderAppointments(providerID uuid.UUID, date time.Time) ([]domain.Appointment, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	var appointments []domain.Appointment
	err := r.db.Where("provider_id = ?", providerID).
		Where("start_time >= ? AND start_time < ?", startOfDay, endOfDay).
		Where("status IN ?", []domain.AppointmentStatus{domain.StatusPending, domain.StatusConfirmed}).
		Find(&appointments).Error

	return appointments, err
}
