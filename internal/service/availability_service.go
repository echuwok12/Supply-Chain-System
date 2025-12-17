package service

import (
	"appointment-booking/internal/domain"
	"appointment-booking/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AvailabilityService struct {
	availRepo *repository.AvailabilityRepository
	apptRepo  *repository.AppointmentRepository
	redis     *redis.Client
}

func NewAvailabilityService(availRepo *repository.AvailabilityRepository, apptRepo *repository.AppointmentRepository, redis *redis.Client) *AvailabilityService {
	return &AvailabilityService{availRepo: availRepo, apptRepo: apptRepo, redis: redis}
}

type SetAvailabilityInput struct {
	DayOfWeek int    `json:"day_of_week" binding:"required,min=0,max=6"`
	StartTime string `json:"start_time" binding:"required"` // "09:00"
	EndTime   string `json:"end_time" binding:"required"`   // "17:00"
}

// Allows a provider to define their schedule
func (s *AvailabilityService) SetAvailability(providerID uuid.UUID, input SetAvailabilityInput) error {
	avail := &domain.Availability{
		ProviderID: providerID,
		DayOfWeek:  input.DayOfWeek,
		StartTime:  input.StartTime,
		EndTime:    input.EndTime,
	}
	return s.availRepo.Save(avail)
}

// Calculates free time slots
func (s *AvailabilityService) GetAvailableSlots(providerID uuid.UUID, dateStr string) ([]time.Time, error) {

	ctx := context.Background()

	// 1. Define Cache Key (e.g., "slots:uuid:2025-10-30")
	cacheKey := fmt.Sprintf("slots:%s:%s", providerID.String(), dateStr)

	// 2. Try Fetching from Redis
	val, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		// HIT: Parse JSON and return
		var slots []time.Time
		if err := json.Unmarshal([]byte(val), &slots); err == nil {
			return slots, nil
		}
	} else if err != redis.Nil {
		// Log error but don't fail; fall back to DB
		fmt.Printf("Redis error: %v\n", err)
	}

	// 1. Parse Date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, errors.New("invalid date format (use YYYY-MM-DD)")
	}

	// 2. Get Working Hours
	dayOfWeek := int(date.Weekday())
	avail, err := s.availRepo.GetByProviderAndDay(providerID, dayOfWeek)
	if err != nil {
		return nil, errors.New("provider not available on this day")
	}

	// 3. Get Existing Appointments
	appointments, err := s.apptRepo.GetProviderAppointments(providerID, date)
	if err != nil {
		return nil, err
	}

	// 4. Algorithm: Generate Slots
	var slots []time.Time

	// Parse "09:00" into actual time for that specific date
	startHour, _ := time.Parse("15:04", avail.StartTime)
	endHour, _ := time.Parse("15:04", avail.EndTime)

	current := time.Date(date.Year(), date.Month(), date.Day(), startHour.Hour(), startHour.Minute(), 0, 0, time.UTC)
	end := time.Date(date.Year(), date.Month(), date.Day(), endHour.Hour(), endHour.Minute(), 0, 0, time.UTC)

	slotDuration := 30 * time.Minute // Hardcoded for now, could be dynamic

	for current.Add(slotDuration).Before(end) || current.Add(slotDuration).Equal(end) {
		slotEnd := current.Add(slotDuration)

		isBooked := false
		for _, appt := range appointments {
			// Check overlap: (StartA < EndB) && (EndA > StartB)
			if current.Before(appt.EndTime) && slotEnd.After(appt.StartTime) {
				isBooked = true
				break
			}
		}

		if !isBooked {
			slots = append(slots, current)
		}

		current = current.Add(slotDuration)
	}

	data, _ := json.Marshal(slots)
	s.redis.Set(ctx, cacheKey, data, 1*time.Minute)

	return slots, nil
}
