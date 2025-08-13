package cleanup

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"user-service/internal/repository"
)

// OTPCleanupService handles cleanup of expired OTP records
type OTPCleanupService struct {
	otpRepo repository.OTPRepository
	interval time.Duration
	maxAge   time.Duration
}

// NewOTPCleanupService creates a new OTP cleanup service
func NewOTPCleanupService(otpRepo repository.OTPRepository, interval, maxAge time.Duration) *OTPCleanupService {
	return &OTPCleanupService{
		otpRepo:  otpRepo,
		interval: interval,
		maxAge:   maxAge,
	}
}

// Start begins the cleanup process in a background goroutine
func (s *OTPCleanupService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	log.Info().
		Dur("interval", s.interval).
		Dur("max_age", s.maxAge).
		Msg("starting OTP cleanup service")

	// Run cleanup immediately on start
	s.cleanup(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping OTP cleanup service")
			return
		case <-ticker.C:
			s.cleanup(ctx)
		}
	}
}

// cleanup removes expired OTP records
func (s *OTPCleanupService) cleanup(ctx context.Context) {
	start := time.Now()
	
	count, err := s.otpRepo.DeleteExpired(ctx, s.maxAge)
	if err != nil {
		log.Error().Err(err).Msg("failed to cleanup expired OTPs")
		return
	}

	duration := time.Since(start)
	log.Info().
		Int("deleted_count", count).
		Dur("duration", duration).
		Msg("OTP cleanup completed")
}
