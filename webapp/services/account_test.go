package services_test

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/argon2"
	"ohmontrouge/webapp/models"
	"ohmontrouge/webapp/services"
)

// stubAccountRepo is a stub AccountRepository for unit tests.
type stubAccountRepo struct {
	account   *models.Account
	err       error
	createdID int64
}

func (s stubAccountRepo) FindByEmail(_ *pop.Connection, _ string) (*models.Account, error) {
	return s.account, s.err
}

func (s stubAccountRepo) GetByID(_ *pop.Connection, _ int64) (*models.Account, error) {
	return s.account, s.err
}

func (s stubAccountRepo) Create(_ *pop.Connection, _, _ string, _ int64) (int64, error) {
	return s.createdID, s.err
}

func (s stubAccountRepo) UpdatePasswordHash(_ *pop.Connection, _ int64, _ string) error {
	return s.err
}

func (s stubAccountRepo) Activate(_ *pop.Connection, _ int64, _ string, _ bool) error {
	return s.err
}

func (s stubAccountRepo) CreatePending(_ *pop.Connection, _ string, _ int64) (int64, error) {
	return s.createdID, s.err
}

func (s stubAccountRepo) UpdateEmail(_ *pop.Connection, _ int64, _ string) error {
	return s.err
}

func (s stubAccountRepo) Delete(_ *pop.Connection, _ int64) error {
	return s.err
}

func (s stubAccountRepo) AnonymizeAccount(_ *pop.Connection, _ int64, _ string) error {
	return s.err
}

// stubRoleRepo is a stub RoleRepository for unit tests.
type stubRoleRepo struct {
	hasActiveHolder bool
	holderErr       error
	roleID          int64
	roleIDErr       error
	assignErr       error
}

func (s stubRoleRepo) HasRole(_ *pop.Connection, _ int64, _ string) (bool, error) {
	return false, nil
}

func (s stubRoleRepo) HasActiveRoleHolder(_ *pop.Connection, _ string) (bool, error) {
	return s.hasActiveHolder, s.holderErr
}

func (s stubRoleRepo) GetIDByName(_ *pop.Connection, _ string) (int64, error) {
	return s.roleID, s.roleIDErr
}

func (s stubRoleRepo) AssignRole(_ *pop.Connection, _, _ int64) error {
	return s.assignErr
}

func (s stubRoleRepo) CountActiveAdmins(_ *pop.Connection) (int, error) {
	return 0, nil
}

func (s stubRoleRepo) RevokeRole(_ *pop.Connection, _, _ int64) error {
	return nil
}

func (s stubRoleRepo) RemoveAllRoles(_ *pop.Connection, _ int64) error {
	return nil
}

// testArgon2Hash generates a PHC-format argon2id hash with minimal parameters for test speed.
func testArgon2Hash(t *testing.T, password string) string {
	t.Helper()
	salt := []byte("testsalt12345678") // 16 bytes, fixed for reproducibility
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 1, 32)
	return fmt.Sprintf("$argon2id$v=%d$m=65536,t=1,p=1$%s$%s",
		argon2.Version,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
}

func TestAuthenticate_AccountNotFound(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{err: sql.ErrNoRows},
	}
	_, err := svc.Authenticate(nil, "nobody@example.com", "password")
	assert.ErrorIs(t, err, services.ErrAccountNotFound)
}

func TestAuthenticate_DBError(t *testing.T) {
	dbErr := errors.New("connection refused")
	svc := services.AccountService{
		Accounts: stubAccountRepo{err: dbErr},
	}
	_, err := svc.Authenticate(nil, "alice@example.com", "password")
	assert.ErrorIs(t, err, dbErr)
}

func TestAuthenticate_InactiveAccount(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{account: &models.Account{
			Email:        nulls.NewString("pending@example.com"),
			PasswordHash: nulls.NewString(testArgon2Hash(t, "correct")),
			Status:       "pending",
		}},
	}
	_, err := svc.Authenticate(nil, "pending@example.com", "correct")
	assert.ErrorIs(t, err, services.ErrInvalidPassword)
}

func TestAuthenticate_NullPasswordHash(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{account: &models.Account{
			Email:        nulls.NewString("active@example.com"),
			PasswordHash: nulls.String{}, // active but not yet activated (defense-in-depth)
			Status:       "active",
		}},
	}
	_, err := svc.Authenticate(nil, "active@example.com", "password")
	assert.ErrorIs(t, err, services.ErrInvalidPassword)
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{account: &models.Account{
			Email:        nulls.NewString("alice@example.com"),
			PasswordHash: nulls.NewString(testArgon2Hash(t, "correct")),
			Status:       "active",
		}},
	}
	_, err := svc.Authenticate(nil, "alice@example.com", "wrong")
	assert.ErrorIs(t, err, services.ErrInvalidPassword)
}

func TestAuthenticate_Success(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{account: &models.Account{
			ID:           1,
			Email:        nulls.NewString("alice@example.com"),
			PasswordHash: nulls.NewString(testArgon2Hash(t, "correct")),
			Status:       "active",
		}},
	}
	got, err := svc.Authenticate(nil, "alice@example.com", "correct")
	assert.NoError(t, err)
	assert.Equal(t, &services.AccountDTO{
		ID:     1,
		Email:  "alice@example.com",
		Status: services.StatusActive,
	}, got)
}

func TestCreateAdmin_Success(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{createdID: 1},
		Roles:    stubRoleRepo{roleID: 1},
	}
	err := svc.CreateAdmin(nil, "admin@example.com", "secret", 1)
	assert.NoError(t, err)
}

func TestCreateAdmin_AlreadyExists(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{},
		Roles:    stubRoleRepo{hasActiveHolder: true},
	}
	err := svc.CreateAdmin(nil, "admin@example.com", "secret", 1)
	assert.ErrorIs(t, err, services.ErrAdminAlreadyExists)
}

func TestResetPassword_Success(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{account: &models.Account{
			ID:     1,
			Email:  nulls.NewString("admin@example.com"),
			Status: "active",
		}},
		Roles: stubRoleRepo{},
	}
	err := svc.ResetPassword(nil, "admin@example.com", "newpassword")
	assert.NoError(t, err)
}

func TestResetPassword_AccountNotFound(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{err: sql.ErrNoRows},
		Roles:    stubRoleRepo{},
	}
	err := svc.ResetPassword(nil, "nobody@example.com", "newpassword")
	assert.ErrorIs(t, err, services.ErrAccountNotFound)
}

func TestResetPassword_AccountNotActive(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{account: &models.Account{
			ID:     1,
			Email:  nulls.NewString("pending@example.com"),
			Status: "pending",
		}},
		Roles: stubRoleRepo{},
	}
	err := svc.ResetPassword(nil, "pending@example.com", "newpassword")
	assert.ErrorIs(t, err, services.ErrAccountNotActive)
}

// --- token stubs ---

type stubInviteTokenRepo struct {
	record           *models.InviteTokenRecord
	findErr          error
	generateErr      error
	markUsedErr      error
	invalidateErr    error
	generatedToken   string
	invalidateCalled bool
}

func (s *stubInviteTokenRepo) Generate(_ *pop.Connection, _ int64, token string, _ time.Time) error {
	s.generatedToken = token
	return s.generateErr
}

func (s *stubInviteTokenRepo) FindByToken(_ *pop.Connection, _ string) (*models.InviteTokenRecord, error) {
	return s.record, s.findErr
}

func (s *stubInviteTokenRepo) MarkUsed(_ *pop.Connection, _ int64) error {
	return s.markUsedErr
}

func (s *stubInviteTokenRepo) InvalidateExisting(_ *pop.Connection, _ int64) error {
	s.invalidateCalled = true
	return s.invalidateErr
}

func (s *stubInviteTokenRepo) FindActiveForAccount(_ *pop.Connection, _ int64) (*models.InviteToken, error) {
	return nil, nil
}

type stubResetTokenRepo struct {
	record           *models.PasswordResetTokenRecord
	findErr          error
	generateErr      error
	markUsedErr      error
	invalidateErr    error
	invalidateCalled bool
}

func (s *stubResetTokenRepo) Generate(_ *pop.Connection, _ int64, _ string, _ time.Time) error {
	return s.generateErr
}

func (s *stubResetTokenRepo) FindByToken(_ *pop.Connection, _ string) (*models.PasswordResetTokenRecord, error) {
	return s.record, s.findErr
}

func (s *stubResetTokenRepo) MarkUsed(_ *pop.Connection, _ int64) error {
	return s.markUsedErr
}

func (s *stubResetTokenRepo) InvalidateExisting(_ *pop.Connection, _ int64) error {
	s.invalidateCalled = true
	return s.invalidateErr
}

func (s *stubResetTokenRepo) FindActiveForAccount(_ *pop.Connection, _ int64) (*models.PasswordResetToken, error) {
	return nil, nil
}

// --- invite token tests ---

func TestGenerateInviteToken_Success(t *testing.T) {
	stub := &stubInviteTokenRepo{}
	svc := services.AccountService{
		Accounts:     stubAccountRepo{},
		Roles:        stubRoleRepo{},
		InviteTokens: stub,
	}
	dto, err := svc.GenerateInviteToken(nil, 1, "https://ohm.test")
	assert.NoError(t, err)
	assert.NotEmpty(t, dto.Token)
	assert.Contains(t, dto.URL, "/invitation/")
	assert.True(t, dto.ExpiresAt.After(time.Now()))
	assert.True(t, stub.invalidateCalled)
}

func TestGenerateInviteToken_InvalidateError(t *testing.T) {
	stub := &stubInviteTokenRepo{invalidateErr: errors.New("db error")}
	svc := services.AccountService{InviteTokens: stub}
	_, err := svc.GenerateInviteToken(nil, 1, "https://ohm.test")
	assert.Error(t, err)
}

func TestValidateInviteToken_Valid(t *testing.T) {
	stub := &stubInviteTokenRepo{record: &models.InviteTokenRecord{
		TokenID:        10,
		AccountID:      5,
		FirstName:      "Alice",
		LastName:       "Dupont",
		Email:          "alice@example.com",
		InstrumentName: "Clarinette",
	}}
	svc := services.AccountService{InviteTokens: stub}
	ctx, err := svc.ValidateInviteToken(nil, "some-token")
	assert.NoError(t, err)
	assert.Equal(t, int64(10), ctx.TokenID)
	assert.Equal(t, "Alice", ctx.FirstName)
	assert.Equal(t, "Clarinette", ctx.InstrumentName)
}

func TestValidateInviteToken_NotFound(t *testing.T) {
	stub := &stubInviteTokenRepo{record: nil}
	svc := services.AccountService{InviteTokens: stub}
	_, err := svc.ValidateInviteToken(nil, "bad-token")
	assert.ErrorIs(t, err, services.ErrInvalidToken)
}

func TestValidateInviteToken_DBError(t *testing.T) {
	stub := &stubInviteTokenRepo{findErr: errors.New("db error")}
	svc := services.AccountService{InviteTokens: stub}
	_, err := svc.ValidateInviteToken(nil, "some-token")
	assert.Error(t, err)
	assert.NotErrorIs(t, err, services.ErrInvalidToken)
}

func TestCompleteInvite_Success(t *testing.T) {
	accounts := stubAccountRepo{}
	tokens := &stubInviteTokenRepo{}
	svc := services.AccountService{Accounts: accounts, InviteTokens: tokens}
	err := svc.CompleteInvite(nil, 10, 5, "hash", true)
	assert.NoError(t, err)
}

func TestCompleteInvite_ActivateError(t *testing.T) {
	accounts := stubAccountRepo{err: errors.New("db error")}
	svc := services.AccountService{Accounts: accounts, InviteTokens: &stubInviteTokenRepo{}}
	err := svc.CompleteInvite(nil, 10, 5, "hash", false)
	assert.Error(t, err)
}

func TestCompleteInvite_MarkUsedError(t *testing.T) {
	accounts := stubAccountRepo{}
	tokens := &stubInviteTokenRepo{markUsedErr: errors.New("db error")}
	svc := services.AccountService{Accounts: accounts, InviteTokens: tokens}
	err := svc.CompleteInvite(nil, 10, 5, "hash", true)
	assert.Error(t, err)
}

// --- password reset token tests ---

func TestGeneratePasswordResetToken_Success(t *testing.T) {
	stub := &stubResetTokenRepo{}
	svc := services.AccountService{ResetTokens: stub}
	dto, err := svc.GeneratePasswordResetToken(nil, 1, "https://ohm.test")
	assert.NoError(t, err)
	assert.NotEmpty(t, dto.Token)
	assert.Contains(t, dto.URL, "/reinitialiser-mot-de-passe/")
	assert.True(t, stub.invalidateCalled)
}

func TestValidatePasswordResetToken_Valid(t *testing.T) {
	stub := &stubResetTokenRepo{record: &models.PasswordResetTokenRecord{
		TokenID:       20,
		AccountID:     7,
		AccountStatus: "active",
	}}
	svc := services.AccountService{ResetTokens: stub}
	ctx, err := svc.ValidatePasswordResetToken(nil, "some-token")
	assert.NoError(t, err)
	assert.Equal(t, int64(20), ctx.TokenID)
	assert.Equal(t, int64(7), ctx.AccountID)
}

func TestValidatePasswordResetToken_NotFound(t *testing.T) {
	svc := services.AccountService{ResetTokens: &stubResetTokenRepo{record: nil}}
	_, err := svc.ValidatePasswordResetToken(nil, "bad-token")
	assert.ErrorIs(t, err, services.ErrInvalidToken)
}

func TestCompletePasswordReset_Success(t *testing.T) {
	svc := services.AccountService{
		Accounts:    stubAccountRepo{},
		ResetTokens: &stubResetTokenRepo{},
	}
	err := svc.CompletePasswordReset(nil, 20, 7, "newhash")
	assert.NoError(t, err)
}

// --- AC-M6: CompleteInvite seeds RSVPs for future events ---

// stubRSVPSeedRepo is a minimal RSVPRepository stub that tracks SeedForAccount calls.
type stubRSVPSeedRepo struct {
	seeded bool
	err    error
}

func (s *stubRSVPSeedRepo) SeedForEvent(_ *pop.Connection, _ int64) error { return s.err }
func (s *stubRSVPSeedRepo) SeedForAccount(_ *pop.Connection, _ int64) error {
	s.seeded = true
	return s.err
}
func (s *stubRSVPSeedRepo) GetByAccountAndEvent(_ *pop.Connection, _, _ int64) (*models.RSVPRow, error) {
	return nil, s.err
}
func (s *stubRSVPSeedRepo) Update(_ *pop.Connection, _ int64, _ string, _ *int64) error { return s.err }
func (s *stubRSVPSeedRepo) DeleteByAccount(_ *pop.Connection, _ int64) error            { return s.err }
func (s *stubRSVPSeedRepo) ClearFieldResponses(_ *pop.Connection, _ int64) error        { return s.err }
func (s *stubRSVPSeedRepo) ListForEvent(_ *pop.Connection, _ int64) ([]models.RSVPListRow, error) {
	return nil, s.err
}
func (s *stubRSVPSeedRepo) ResetYesRSVPs(_ *pop.Connection, _ int64) error    { return s.err }
func (s *stubRSVPSeedRepo) ClearInstruments(_ *pop.Connection, _ int64) error { return s.err }
func (s *stubRSVPSeedRepo) AddFieldResponse(_ *pop.Connection, _, _ int64, _ string) error {
	return s.err
}
func (s *stubRSVPSeedRepo) ListFieldResponses(_ *pop.Connection, _ int64) ([]models.RSVPFieldResponseRow, error) {
	return nil, s.err
}

func TestCompleteInvite_SeedsRSVPs(t *testing.T) {
	rsvps := &stubRSVPSeedRepo{}
	svc := services.AccountService{
		Accounts:     stubAccountRepo{},
		InviteTokens: &stubInviteTokenRepo{},
		Events:       rsvps,
	}
	err := svc.CompleteInvite(nil, 10, 5, "hash", true)
	assert.NoError(t, err)
	assert.True(t, rsvps.seeded, "SeedForAccount should be called on invite completion")
}

func TestCompleteInvite_NilEvents_DoesNotPanic(t *testing.T) {
	svc := services.AccountService{
		Accounts:     stubAccountRepo{},
		InviteTokens: &stubInviteTokenRepo{},
		Events:       nil,
	}
	err := svc.CompleteInvite(nil, 10, 5, "hash", true)
	assert.NoError(t, err)
}

// --- password strength tests ---

func TestValidatePasswordStrength_TooShort(t *testing.T) {
	err := services.ValidatePasswordStrength("Short1!", "Short1!")
	assert.Error(t, err)
}

func TestValidatePasswordStrength_MissingUppercase(t *testing.T) {
	err := services.ValidatePasswordStrength("alllowercase1!longenough", "alllowercase1!longenough")
	assert.Error(t, err)
}

func TestValidatePasswordStrength_MissingDigit(t *testing.T) {
	err := services.ValidatePasswordStrength("NoDigitHereAtAllLong!", "NoDigitHereAtAllLong!")
	assert.Error(t, err)
}

func TestValidatePasswordStrength_MissingSpecial(t *testing.T) {
	err := services.ValidatePasswordStrength("NoSpecialChar1234567890", "NoSpecialChar1234567890")
	assert.Error(t, err)
}

func TestValidatePasswordStrength_Mismatch(t *testing.T) {
	err := services.ValidatePasswordStrength("Valid1!Password_long", "different")
	assert.Error(t, err)
}

func TestValidatePasswordStrength_Valid(t *testing.T) {
	pw := "ValidPassword1!ExtraLong"
	err := services.ValidatePasswordStrength(pw, pw)
	assert.NoError(t, err)
}
