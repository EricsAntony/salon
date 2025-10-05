package validation

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/EricsAntony/salon/salon-shared/errors"
	"github.com/google/uuid"
)

var (
	// nameRegex validates name format (letters, spaces, hyphens, apostrophes, dots)
	nameRegex = regexp.MustCompile(`^[a-zA-Z\s\-'\.]+$`)
	
	// urlRegex validates URL format
	urlRegex = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
)

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return nil // Email is optional in most cases
	}
	
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.NewValidationError("email", "invalid format")
	}
	
	return nil
}

// ValidateRequiredEmail validates required email format
func ValidateRequiredEmail(email string) error {
	if email == "" {
		return errors.NewValidationError("email", "is required")
	}
	
	return ValidateEmail(email)
}

// ValidateName validates name fields
func ValidateName(name, fieldName string) error {
	if strings.TrimSpace(name) == "" {
		return errors.NewValidationError(fieldName, "is required")
	}
	
	if utf8.RuneCountInString(name) < 2 {
		return errors.NewValidationError(fieldName, "must be at least 2 characters")
	}
	
	if utf8.RuneCountInString(name) > 100 {
		return errors.NewValidationError(fieldName, "must be less than 100 characters")
	}
	
	if !nameRegex.MatchString(name) {
		return errors.NewValidationError(fieldName, "contains invalid characters")
	}
	
	return nil
}

// ValidateDescription validates description fields
func ValidateDescription(desc *string, fieldName string, maxLength int) error {
	if desc == nil {
		return nil
	}
	
	if utf8.RuneCountInString(*desc) > maxLength {
		return errors.NewValidationError(fieldName, fmt.Sprintf("must be less than %d characters", maxLength))
	}
	
	return nil
}

// ValidatePrice validates price values
func ValidatePrice(price float64) error {
	if price < 0 {
		return errors.NewValidationError("price", "must be non-negative")
	}
	
	if price > 999999.99 {
		return errors.NewValidationError("price", "exceeds maximum allowed value")
	}
	
	return nil
}

// ValidateDuration validates duration in minutes
func ValidateDuration(duration int) error {
	if duration <= 0 {
		return errors.NewValidationError("duration_minutes", "must be positive")
	}
	
	if duration > 1440 { // 24 hours
		return errors.NewValidationError("duration_minutes", "cannot exceed 24 hours")
	}
	
	return nil
}

// ValidateTaxRate validates tax rate percentage
func ValidateTaxRate(rate float64) error {
	if rate < 0 {
		return errors.NewValidationError("tax_rate", "must be non-negative")
	}
	
	if rate > 100 {
		return errors.NewValidationError("tax_rate", "cannot exceed 100%")
	}
	
	return nil
}

// ValidateURL validates URL format
func ValidateURL(url *string, fieldName string) error {
	if url == nil || *url == "" {
		return nil
	}
	
	if !urlRegex.MatchString(*url) {
		return errors.NewValidationError(fieldName, "invalid URL format")
	}
	
	return nil
}

// ValidateStringArray validates string arrays
func ValidateStringArray(arr []string, fieldName string, maxItems, maxItemLength int) error {
	if len(arr) > maxItems {
		return errors.NewValidationError(fieldName, fmt.Sprintf("cannot have more than %d items", maxItems))
	}
	
	for i, item := range arr {
		if utf8.RuneCountInString(item) > maxItemLength {
			return errors.NewValidationError(
				fmt.Sprintf("%s[%d]", fieldName, i),
				fmt.Sprintf("must be less than %d characters", maxItemLength),
			)
		}
	}
	
	return nil
}

// ValidateUUID validates UUID format
func ValidateUUID(id, fieldName string) error {
	if strings.TrimSpace(id) == "" {
		return errors.NewValidationError(fieldName, "is required")
	}
	
	if _, err := uuid.Parse(id); err != nil {
		return errors.NewValidationError(fieldName, "must be a valid UUID")
	}
	
	return nil
}

// ValidateRequired validates that a string field is not empty
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return errors.NewValidationError(fieldName, "is required")
	}
	return nil
}

// ValidateLength validates string length constraints
func ValidateLength(value, fieldName string, minLength, maxLength int) error {
	length := utf8.RuneCountInString(value)
	
	if length < minLength {
		return errors.NewValidationError(fieldName, fmt.Sprintf("must be at least %d characters", minLength))
	}
	
	if length > maxLength {
		return errors.NewValidationError(fieldName, fmt.Sprintf("must be less than %d characters", maxLength))
	}
	
	return nil
}
