package repository

import (
	"appointment-booking/internal/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AppointmentRepository struct {
	db *gorm.DB
}

// Struct to hold raw SQL results
type AppointmentStats struct {
	Status string
	Count  int64
}

type ProviderLeaderboard struct {
	ProviderID string
	Name       string
	Total      int64
}

func NewAppointmentRepository(db *gorm.DB) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

func (r *AppointmentRepository) Create(tx *gorm.DB, appointment *domain.Appointment) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(appointment).Error
}

func (r *AppointmentRepository) HasOverlap(providerID uuid.UUID, start, end time.Time) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Appointment{}).
		Where("provider_id = ?", providerID).
		Where("status IN ?", []domain.AppointmentStatus{domain.StatusPending, domain.StatusConfirmed}).
		Where("start_time < ? AND end_time > ?", end, start).
		Count(&count).Error

	return count > 0, err
}

func (r *AppointmentRepository) BeginTx() *gorm.DB {
	return r.db.Begin()
}

// Fetches an appointment (needed to check ownership/status before modifying)
func (r *AppointmentRepository) FindByID(id uuid.UUID) (*domain.Appointment, error) {
	var appt domain.Appointment
	err := r.db.First(&appt, "id = ?", id).Error
	return &appt, err
}

// Saves changes to an existing appointment
func (r *AppointmentRepository) Update(appt *domain.Appointment) error {
	return r.db.Save(appt).Error
}

// GetStatsByDateRange returns count of appointments grouped by status
func (r *AppointmentRepository) GetStatsByDateRange(start, end time.Time) ([]AppointmentStats, error) {
	var stats []AppointmentStats

	// SQL: SELECT status, count(*) FROM appointments WHERE start_time BETWEEN ? AND ? GROUP BY status
	err := r.db.Model(&domain.Appointment{}).
		Select("status, count(*) as count").
		Where("start_time BETWEEN ? AND ?", start, end).
		Group("status").
		Scan(&stats).Error

	return stats, err
}

// GetTopProviders returns providers with the most completed appointments
func (r *AppointmentRepository) GetTopProviders(limit int) ([]ProviderLeaderboard, error) {
	var results []ProviderLeaderboard

	// SQL: JOIN users on appointments to get names, Count, Sort DESC
	err := r.db.Table("appointments").
		Select("users.id as provider_id, users.name, count(appointments.id) as total").
		Joins("JOIN users ON users.id = appointments.provider_id").
		Where("appointments.status = ?", domain.StatusCompleted).
		Group("users.id, users.name").
		Order("total desc").
		Limit(limit).
		Scan(&results).Error

	return results, err
}
