package validation

import (
	"regexp"
	"strings"

	"github.com/EricsAntony/salon/salon-shared/errors"
)

var (
	// phoneRegex validates basic phone number format
	phoneRegex = regexp.MustCompile(`^[0-9+\-\s\(\)]+$`)
	
	// digitOnlyRegex extracts only digits and plus sign
	digitOnlyRegex = regexp.MustCompile(`[^0-9+]`)
)

// NormalizePhone standardizes phone number format
func NormalizePhone(phone string) string {
	if phone == "" {
		return ""
	}
	
	// Remove all non-digit characters except +
	cleaned := strings.TrimSpace(phone)
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, ".", "")
	
	if cleaned == "" {
		return ""
	}
	
	// Handle different phone number formats
	if strings.HasPrefix(cleaned, "+") {
		return cleaned
	}
	
	// Indian number handling
	if len(cleaned) == 10 {
		return "+91" + cleaned
	}
	
	if strings.HasPrefix(cleaned, "0") && len(cleaned) == 11 {
		return "+91" + cleaned[1:]
	}
	
	if strings.HasPrefix(cleaned, "91") && len(cleaned) == 12 {
		return "+" + cleaned
	}
	
	// Return as-is if we can't determine format
	return cleaned
}

// ValidatePhone validates phone number format and returns normalized version
func ValidatePhone(phone string) (string, error) {
	if phone == "" {
		return "", errors.NewValidationError("phone_number", "is required")
	}
	
	// Basic format check
	if !phoneRegex.MatchString(phone) {
		return "", errors.NewValidationError("phone_number", "contains invalid characters")
	}
	
	normalized := NormalizePhone(phone)
	if normalized == "" {
		return "", errors.NewValidationError("phone_number", "invalid format")
	}
	
	// Validate normalized phone
	if strings.HasPrefix(normalized, "+") {
		if len(normalized) < 10 || len(normalized) > 15 {
			return "", errors.NewValidationError("phone_number", "invalid length")
		}
		
		// Check if all characters after + are digits
		for _, char := range normalized[1:] {
			if char < '0' || char > '9' {
				return "", errors.NewValidationError("phone_number", "invalid format")
			}
		}
	} else {
		// For numbers without country code
		if len(normalized) < 7 || len(normalized) > 15 {
			return "", errors.NewValidationError("phone_number", "invalid length")
		}
		
		// Check if all characters are digits
		for _, char := range normalized {
			if char < '0' || char > '9' {
				return "", errors.NewValidationError("phone_number", "invalid format")
			}
		}
	}
	
	return normalized, nil
}

// IsValidPhoneFormat checks if phone number has valid format without normalization
func IsValidPhoneFormat(phone string) bool {
	_, err := ValidatePhone(phone)
	return err == nil
}
