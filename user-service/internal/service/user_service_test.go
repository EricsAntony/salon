package service

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/salon-shared/auth"
	"github.com/your-org/salon-shared/config"
	"github.com/your-org/salon-shared/models"
)

type fakeUserRepo struct{ u *models.User; deleted string; updated bool }

func (f *fakeUserRepo) Create(ctx context.Context, u *models.User) error { f.u = u; return nil }
func (f *fakeUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	if f.u != nil && f.u.ID == id { return f.u, nil }
	return nil, nil
}
func (f *fakeUserRepo) GetByPhone(ctx context.Context, phone string) (*models.User, error) { return nil, nil }
func (f *fakeUserRepo) Update(ctx context.Context, u *models.User) error { f.updated = true; f.u = u; return nil }
func (f *fakeUserRepo) Delete(ctx context.Context, id string) error { f.deleted = id; return nil }

type fakeOTPRepo struct{}
func (f *fakeOTPRepo) Create(ctx context.Context, phone, codeHash string, expiresAt time.Time) error { return nil }
func (f *fakeOTPRepo) GetLatest(ctx context.Context, phone string) (*models.OTP, error) { return nil, nil }
func (f *fakeOTPRepo) IncrementAttempts(ctx context.Context, id int64) error { return nil }

type fakeTokenRepo struct{ valid bool; revokedAll bool; revokedExact bool; saved bool }
func (f *fakeTokenRepo) Save(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error { f.saved = true; return nil }
func (f *fakeTokenRepo) RevokeAllForUser(ctx context.Context, userID string) error { f.revokedAll = true; return nil }
func (f *fakeTokenRepo) IsValid(ctx context.Context, userID, tokenHash string, now time.Time) (bool, error) { return f.valid, nil }
func (f *fakeTokenRepo) RevokeExact(ctx context.Context, userID, tokenHash string) error { f.revokedExact = true; return nil }

func makeServiceForTest() (*userService, *fakeUserRepo, *fakeTokenRepo) {
	cfg := &config.Config{}
	cfg.JWT.AccessSecret = "test-access"
	cfg.JWT.RefreshSecret = "test-refresh"
	cfg.JWT.AccessTTLMinutes = 15
	cfg.JWT.RefreshTTLDays = 7
	jwt := auth.NewJWTManager(cfg)
	ur := &fakeUserRepo{u: &models.User{ID: "uid-1", Name: "Alice"}}
	tr := &fakeTokenRepo{valid: true}
	or := &fakeOTPRepo{}
	s := &userService{users: ur, tokens: tr, otps: or, jwt: jwt, cfg: cfg}
	return s, ur, tr
}

func TestRefresh_Success(t *testing.T) {
	s, _, tr := makeServiceForTest()
	// issue a refresh token for uid-1
	rt, _, err := s.jwt.GenerateRefreshToken("uid-1")
	if err != nil { t.Fatal(err) }
	access, newRT, err := s.Refresh(context.Background(), rt)
	if err != nil { t.Fatalf("refresh failed: %v", err) }
	if access == "" || newRT == "" { t.Fatalf("expected tokens, got empty") }
	if !tr.revokedExact || !tr.saved { t.Fatalf("expected token rotation (revokedExact && saved)") }
}

func TestRevoke_All(t *testing.T) {
	s, _, tr := makeServiceForTest()
	if err := s.Revoke(context.Background(), "uid-1"); err != nil { t.Fatalf("revoke: %v", err) }
	if !tr.revokedAll { t.Fatalf("expected revokedAll = true") }
}

func TestDeleteUser_SelfOnly(t *testing.T) {
	s, ur, _ := makeServiceForTest()
	if err := s.DeleteUser(context.Background(), "uid-1", "uid-2"); err == nil { t.Fatalf("expected forbidden error") }
	if err := s.DeleteUser(context.Background(), "uid-1", "uid-1"); err != nil { t.Fatalf("delete self: %v", err) }
	if ur.deleted != "uid-1" { t.Fatalf("expected delete uid-1") }
}

func TestUpdateUser(t *testing.T) {
	s, ur, _ := makeServiceForTest()
	name := "Bob"
	email := "bob@example.com"
	gender := "male"
	loc := "NYC"
	u, err := s.UpdateUser(context.Background(), "uid-1", &name, &gender, &email, &loc)
	if err != nil { t.Fatalf("update: %v", err) }
	if !ur.updated { t.Fatalf("expected user updated") }
	if u.Name != name || (u.Email == nil || *u.Email != email) || (u.Location == nil || *u.Location != loc) { t.Fatalf("fields not updated") }
}
