package service

import (
	"appointment-booking/internal/domain"
	"appointment-booking/internal/repository"
	"appointment-booking/internal/websocket"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AppointmentService struct {
	repo      *repository.AppointmentRepository
	notifier  *NotificationService
	wsHandler *websocket.Handler
	redis     *redis.Client
}

func NewAppointmentService(repo *repository.AppointmentRepository, notifier *NotificationService, ws *websocket.Handler, redis *redis.Client) *AppointmentService {
	return &AppointmentService{
		repo:      repo,
		notifier:  notifier,
		wsHandler: ws,
		redis:     redis,
	}
}

type BookingInput struct {
	ProviderID  string    `json:"provider_id" binding:"required"`
	ServiceType string    `json:"service_type" binding:"required"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
}

func (s *AppointmentService) BookAppointment(customerID uuid.UUID, input BookingInput) (*domain.Appointment, error) {
	// 1. Validate Time
	if input.EndTime.Before(input.StartTime) {
		return nil, errors.New("end time must be after start time")
	}
	if input.StartTime.Before(time.Now()) {
		return nil, errors.New("cannot book appointments in the past")
	}

	providerUUID, err := uuid.Parse(input.ProviderID)
	if err != nil {
		return nil, errors.New("invalid provider ID")
	}

	// 2. Start Transaction
	tx := s.repo.BeginTx()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 3. Check for Overlaps
	hasOverlap, err := s.repo.HasOverlap(providerUUID, input.StartTime, input.EndTime)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if hasOverlap {
		tx.Rollback()
		return nil, errors.New("time slot is not available")
	}

	// 4. Create Appointment
	appointment := &domain.Appointment{
		CustomerID:  customerID,
		ProviderID:  providerUUID,
		ServiceType: input.ServiceType,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		Status:      domain.StatusPending,
	}

	if err := s.repo.Create(tx, appointment); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 5. Commit Transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	s.notifier.SendAsync(customerID, "Your appointment is confirmed!")
	s.notifier.SendAsync(appointment.ProviderID, "You have a new booking!")

	// 2. Push Real-time Update (Sync/Non-blocking via channel)
	s.wsHandler.Broadcast(map[string]interface{}{
		"event":       "new_booking",
		"provider_id": appointment.ProviderID,
		"slot":        appointment.StartTime,
	})

	// Invalidate Cache for that Provider + Date
	dateStr := input.StartTime.Format("2006-01-02")
	cacheKey := fmt.Sprintf("slots:%s:%s", appointment.ProviderID.String(), dateStr)

	// We ignore errors here; if Redis is down, it just means cache expires naturally later
	s.redis.Del(context.Background(), cacheKey)

	return appointment, nil
}

func (s *AppointmentService) CancelAppointment(appointmentID, userID uuid.UUID) error {
	// 1. Fetch Appointment
	appt, err := s.repo.FindByID(appointmentID)
	if err != nil {
		return errors.New("appointment not found")
	}

	// 2. Authorization Check (Is this the user's appointment?)
	// In a real app, Providers/Admins should also be able to cancel.
	if appt.CustomerID != userID && appt.ProviderID != userID {
		return errors.New("unauthorized to modify this appointment")
	}

	// 3. State Validation
	if appt.Status == domain.StatusCompleted {
		return errors.New("cannot cancel a completed appointment")
	}
	if appt.Status == domain.StatusCancelled {
		return errors.New("appointment is already cancelled")
	}

	// 4. Update Status
	appt.Status = domain.StatusCancelled
	return s.repo.Update(appt)
}

func (s *AppointmentService) RescheduleAppointment(appointmentID, userID uuid.UUID, newStart, newEnd time.Time) error {
	// 1. Fetch & Validate Ownership
	appt, err := s.repo.FindByID(appointmentID)
	if err != nil {
		return errors.New("appointment not found")
	}

	if appt.CustomerID != userID {
		return errors.New("unauthorized")
	}

	if appt.Status == domain.StatusCancelled || appt.Status == domain.StatusCompleted {
		return errors.New("cannot reschedule completed or cancelled appointments")
	}

	// 2. Validate New Time
	if newEnd.Before(newStart) {
		return errors.New("invalid time range")
	}

	// 3. Check Availability for the NEW time
	// Note: HasOverlap checks ALL appointments. It might flag THIS appointment as a conflict
	// if the times overlap slightly.
	// ideally, we should exclude the current appointment ID from the overlap check.
	// For this phase, we assume the user is moving to a completely different slot.
	hasOverlap, err := s.repo.HasOverlap(appt.ProviderID, newStart, newEnd)
	if err != nil {
		return err
	}
	if hasOverlap {
		return errors.New("new time slot is not available")
	}

	// 4. Update
	appt.StartTime = newStart
	appt.EndTime = newEnd
	appt.Status = domain.StatusConfirmed // Auto-confirm on reschedule? Business decision.

	return s.repo.Update(appt)
}
