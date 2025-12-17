package service

import (
	"log"
	"time"

	"github.com/google/uuid"
)

type NotificationPayload struct {
	UserID  uuid.UUID
	Message string
}

type NotificationService struct {
	// Buffered channel to hold messages
	notifyChan chan NotificationPayload
}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		// Buffer of 100: prevents blocking if bursts of requests come in
		notifyChan: make(chan NotificationPayload, 100),
	}
}

// SendAsync pushes the message to the queue and returns immediately
func (s *NotificationService) SendAsync(userID uuid.UUID, message string) {
	s.notifyChan <- NotificationPayload{UserID: userID, Message: message}
}

// StartWorker is the background process that actually sends the emails
func (s *NotificationService) StartWorker() {
	go func() {
		for payload := range s.notifyChan {
			// Simulate slow email server (e.g., SMTP latency)
			time.Sleep(2 * time.Second)

			// In a real app, you would call SendGrid/AWS SES here
			log.Printf("📧 [Email Sent] To User: %s | Body: %s", payload.UserID, payload.Message)
		}
	}()
}
