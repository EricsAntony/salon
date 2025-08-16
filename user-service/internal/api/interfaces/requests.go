package interfaces

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"

	"github.com/EricsAntony/salon/salon-shared/utils"
)

var (
	otpRe   = regexp.MustCompile(`^\d{6}$`)
	// Allow letters, spaces, hyphen, apostrophe, dot; length 2-100
	nameRe  = regexp.MustCompile(`^[A-Za-z][A-Za-z .'-]{1,99}$`)
)

// ReqOTP represents the request body for requesting an OTP
// POST /otp/request
type ReqOTP struct {
	PhoneNumber string `json:"phone_number"`
}

func (r *ReqOTP) ValidateStrict() error {
	if r == nil { return errors.New("invalid request") }
	r.PhoneNumber = strings.TrimSpace(r.PhoneNumber)
	if r.PhoneNumber == "" { return errors.New("phone number required") }
	if !utils.ValidPhone(r.PhoneNumber) { return errors.New("invalid phone number format") }
	return nil
}

// RegisterReq represents the request body for registering a new user
// POST /user/register
type RegisterReq struct {
	PhoneNumber string  `json:"phone_number"`
	Name        string  `json:"name"`
	Gender      string  `json:"gender"`
	Email       *string `json:"email"`
	Location    *string `json:"location"`
	OTP         string  `json:"otp"`
}

func (r *RegisterReq) ValidateStrict() error {
	if r == nil { return errors.New("invalid request") }
	// Trim all
	r.PhoneNumber = strings.TrimSpace(r.PhoneNumber)
	r.Name = strings.TrimSpace(r.Name)
	r.Gender = strings.ToLower(strings.TrimSpace(r.Gender))
	r.OTP = strings.TrimSpace(r.OTP)
	if r.Email != nil { v := strings.TrimSpace(*r.Email); r.Email = &v }
	if r.Location != nil { v := strings.TrimSpace(*r.Location); r.Location = &v }

	// Requireds
	if r.PhoneNumber == "" || r.Name == "" || r.Gender == "" || r.OTP == "" {
		return errors.New("missing required fields")
	}
	// Phone format
	if !utils.ValidPhone(r.PhoneNumber) {
		return errors.New("invalid phone number format")
	}
	// Name constraints
	if len(r.Name) < 2 || len(r.Name) > 100 || !nameRe.MatchString(r.Name) {
		return errors.New("invalid name")
	}
	// Gender enum
	switch r.Gender {
	case "male", "female", "other":
		// ok
	default:
		return errors.New("invalid gender")
	}
	// Email optional but must be valid
	if r.Email != nil && *r.Email != "" {
		if _, err := mail.ParseAddress(*r.Email); err != nil {
			return errors.New("invalid email")
		}
	}
	// Location optional; length cap
	if r.Location != nil && len(*r.Location) > 255 {
		return errors.New("location too long")
	}
	// OTP 6 digits
	if !otpRe.MatchString(r.OTP) {
		return errors.New("invalid otp format")
	}
	return nil
}

// AuthReq represents the request body for authenticating using OTP
// POST /user/authenticate
type AuthReq struct {
	PhoneNumber string `json:"phone_number"`
	OTP         string `json:"otp"`
}

func (r *AuthReq) ValidateStrict() error {
	if r == nil { return errors.New("invalid request") }
	r.PhoneNumber = strings.TrimSpace(r.PhoneNumber)
	r.OTP = strings.TrimSpace(r.OTP)
	if r.PhoneNumber == "" || r.OTP == "" {
		return errors.New("missing phone or otp")
	}
	if !utils.ValidPhone(r.PhoneNumber) {
		return errors.New("invalid phone number format")
	}
	if !otpRe.MatchString(r.OTP) {
		return errors.New("invalid otp format")
	}
	return nil
}

// UpdateReq represents the request body for updating a user profile
// PUT /users/{id}
type UpdateReq struct {
	Name     *string `json:"name"`
	Gender   *string `json:"gender"`
	Email    *string `json:"email"`
	Location *string `json:"location"`
}

func (r *UpdateReq) ValidateStrict() error {
	if r == nil { return errors.New("invalid request") }
	if r.Name == nil && r.Gender == nil && r.Email == nil && r.Location == nil {
		return errors.New("no fields to update")
	}
	if r.Name != nil {
		v := strings.TrimSpace(*r.Name)
		if len(v) < 2 || len(v) > 100 || !nameRe.MatchString(v) {
			return errors.New("invalid name")
		}
		r.Name = &v
	}
	if r.Gender != nil {
		v := strings.ToLower(strings.TrimSpace(*r.Gender))
		switch v {
		case "male", "female", "other":
			// ok
		default:
			return errors.New("invalid gender")
		}
		r.Gender = &v
	}
	if r.Email != nil {
		v := strings.TrimSpace(*r.Email)
		if v != "" {
			if _, err := mail.ParseAddress(v); err != nil {
				return errors.New("invalid email")
			}
		}
		r.Email = &v
	}
	if r.Location != nil {
		v := strings.TrimSpace(*r.Location)
		if len(v) > 255 {
			return errors.New("location too long")
		}
		r.Location = &v
	}
	return nil
}
