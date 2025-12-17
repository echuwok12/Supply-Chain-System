package service

import (
	"appointment-booking/internal/repository"
	"time"
)

type ReportService struct {
	repo *repository.AppointmentRepository
}

func NewReportService(repo *repository.AppointmentRepository) *ReportService {
	return &ReportService{repo: repo}
}

type DashboardData struct {
	TotalAppointments int64                            `json:"total_appointments"`
	Breakdown         map[string]int64                 `json:"breakdown"`
	TopProviders      []repository.ProviderLeaderboard `json:"top_providers"`
}

func (s *ReportService) GetDashboardStats() (*DashboardData, error) {
	// 1. Define Range (e.g., Last 30 Days)
	end := time.Now()
	start := end.AddDate(0, 0, -30)

	// 2. Fetch Stats
	stats, err := s.repo.GetStatsByDateRange(start, end)
	if err != nil {
		return nil, err
	}

	// 3. Fetch Leaderboard
	topProviders, err := s.repo.GetTopProviders(5)
	if err != nil {
		return nil, err
	}

	// 4. Format Data
	breakdown := make(map[string]int64)
	var total int64 = 0

	for _, s := range stats {
		breakdown[s.Status] = s.Count
		total += s.Count
	}

	return &DashboardData{
		TotalAppointments: total,
		Breakdown:         breakdown,
		TopProviders:      topProviders,
	}, nil
}
