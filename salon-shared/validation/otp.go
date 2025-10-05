package validation

import (
	"crypto/rand"
	"fmt"
	"regexp"

	"github.com/EricsAntony/salon/salon-shared/errors"
)

var (
	// otpRegex validates OTP format (6 digits)
	otpRegex = regexp.MustCompile(`^[0-9]{6}$`)
)

// GenerateOTP generates a secure 6-digit OTP
func GenerateOTP() (string, error) {
	// Generate 4 random bytes
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}
	
	// Convert to 6-digit number
	num := uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])
	otp := fmt.Sprintf("%06d", num%1000000)
	
	return otp, nil
}

// ValidateOTP validates OTP format
func ValidateOTP(otp string) error {
	if otp == "" {
		return errors.NewValidationError("otp", "is required")
	}
	
	if len(otp) != 6 {
		return errors.NewValidationError("otp", "must be 6 digits")
	}
	
	if !otpRegex.MatchString(otp) {
		return errors.NewValidationError("otp", "must contain only digits")
	}
	
	return nil
}

// IsValidOTPFormat checks if OTP has valid format
func IsValidOTPFormat(otp string) bool {
	return ValidateOTP(otp) == nil
}
