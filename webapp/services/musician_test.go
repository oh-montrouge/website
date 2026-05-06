package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/stretchr/testify/assert"
	"ohmontrouge/webapp/models"
	"ohmontrouge/webapp/services"
)

var errDBFailure = errors.New("db failure")

// stubMembershipRepo is a stub MembershipRepository for unit tests.
type stubMembershipRepo struct {
	profileRow    *models.MusicianProfileRow
	profileErr    error
	setProfileErr error
	updateErr     error
	listRows      []models.MusicianListRow
	listErr       error
	retentionRows []models.RetentionRow
	retentionErr  error
	clearErr      error
	withdrawErr   error
	toggleErr     error
}

func (s *stubMembershipRepo) GetProfile(_ *pop.Connection, _ int64) (*models.MusicianProfileRow, error) {
	return s.profileRow, s.profileErr
}

func (s *stubMembershipRepo) SetProfile(_ *pop.Connection, _ int64, _, _ string, _ *time.Time, _ string) error {
	return s.setProfileErr
}

func (s *stubMembershipRepo) UpdateProfile(_ *pop.Connection, _ int64, _, _, _ string, _ int64, _ *time.Time, _, _, _ string) error {
	return s.updateErr
}

func (s *stubMembershipRepo) ListNonAnonymized(_ *pop.Connection) ([]models.MusicianListRow, error) {
	return s.listRows, s.listErr
}

func (s *stubMembershipRepo) ListForRetentionReview(_ *pop.Connection) ([]models.RetentionRow, error) {
	return s.retentionRows, s.retentionErr
}

func (s *stubMembershipRepo) ClearMembershipFields(_ *pop.Connection, _ int64) error {
	return s.clearErr
}

func (s *stubMembershipRepo) WithdrawConsent(_ *pop.Connection, _ int64) error {
	return s.withdrawErr
}

func (s *stubMembershipRepo) ToggleProcessingRestriction(_ *pop.Connection, _ int64) error {
	return s.toggleErr
}

// --- AC-M2: Under-15 parental consent rule ---

func TestSetInitialProfile_Under15_WithConsent_Succeeds(t *testing.T) {
	stub := &stubMembershipRepo{}
	svc := services.MembershipService{Membership: stub}

	youngBirth := time.Now().AddDate(-10, 0, 0) // 10 years old — under 15
	err := svc.SetInitialProfile(nil, 1, "Léa", "Martin", &youngBirth, "https://docs.example.com/consent.pdf")
	assert.NoError(t, err)
}

func TestSetInitialProfile_Under15_NoConsent_ReturnsError(t *testing.T) {
	stub := &stubMembershipRepo{}
	svc := services.MembershipService{Membership: stub}

	youngBirth := time.Now().AddDate(-10, 0, 0) // 10 years old — under 15
	err := svc.SetInitialProfile(nil, 1, "Léa", "Martin", &youngBirth, "")
	assert.ErrorIs(t, err, services.ErrParentalConsentRequired)
}

func TestSetInitialProfile_Exactly15_NoConsent_Succeeds(t *testing.T) {
	stub := &stubMembershipRepo{}
	svc := services.MembershipService{Membership: stub}

	birth := time.Now().AddDate(-15, 0, -1) // 15 years and 1 day — not under 15
	err := svc.SetInitialProfile(nil, 1, "Paul", "Dupont", &birth, "")
	assert.NoError(t, err)
}

func TestSetInitialProfile_NoBirthDate_NoConsent_Succeeds(t *testing.T) {
	stub := &stubMembershipRepo{}
	svc := services.MembershipService{Membership: stub}

	err := svc.SetInitialProfile(nil, 1, "Marc", "Leroy", nil, "")
	assert.NoError(t, err)
}

func TestUpdateProfile_Under15_NoConsent_ReturnsError(t *testing.T) {
	stub := &stubMembershipRepo{}
	svc := services.MembershipService{Membership: stub}

	youngBirth := time.Now().AddDate(-12, 0, 0)
	err := svc.UpdateProfile(nil, 1, "Alice", "Dupont", "alice@example.com", 1, &youngBirth, "", "", "")
	assert.ErrorIs(t, err, services.ErrParentalConsentRequired)
}

func TestUpdateProfile_Over15_NoConsent_Succeeds(t *testing.T) {
	stub := &stubMembershipRepo{}
	svc := services.MembershipService{Membership: stub}

	adultBirth := time.Now().AddDate(-20, 0, 0)
	err := svc.UpdateProfile(nil, 1, "Alice", "Dupont", "alice@example.com", 1, &adultBirth, "", "", "")
	assert.NoError(t, err)
}

// --- AC-M6: Consent withdrawal ---

func TestConsentWithdrawal_CallsWithdrawConsent(t *testing.T) {
	stub := &stubMembershipRepo{}
	svc := services.MembershipService{Membership: stub}

	err := svc.ConsentWithdrawal(nil, 5)
	assert.NoError(t, err)
}

func TestConsentWithdrawal_PropagatesError(t *testing.T) {
	stub := &stubMembershipRepo{withdrawErr: errDBFailure}
	svc := services.MembershipService{Membership: stub}

	err := svc.ConsentWithdrawal(nil, 5)
	assert.ErrorIs(t, err, errDBFailure)
}

// --- ListNonAnonymized mapping ---

func TestListNonAnonymized_MapsRows(t *testing.T) {
	stub := &stubMembershipRepo{
		listRows: []models.MusicianListRow{
			{ID: 1, InstrumentName: "Clarinette", Status: "active", IsAdmin: true},
		},
	}
	svc := services.MembershipService{Membership: stub}

	summaries, err := svc.ListNonAnonymized(nil)
	assert.NoError(t, err)
	assert.Len(t, summaries, 1)
	assert.Equal(t, int64(1), summaries[0].AccountID)
	assert.Equal(t, "Clarinette", summaries[0].MainInstrumentName)
	assert.True(t, summaries[0].IsAdmin)
}

func TestListNonAnonymized_Error(t *testing.T) {
	stub := &stubMembershipRepo{listErr: errDBFailure}
	svc := services.MembershipService{Membership: stub}

	summaries, err := svc.ListNonAnonymized(nil)
	assert.Nil(t, summaries)
	assert.ErrorIs(t, err, errDBFailure)
}

// --- GetProfile ---

func TestGetProfile_Success_WithBirthDate(t *testing.T) {
	bt := time.Now().AddDate(-20, 0, 0)
	stub := &stubMembershipRepo{
		profileRow: &models.MusicianProfileRow{
			ID:             1,
			FirstName:      nulls.NewString("Alice"),
			LastName:       nulls.NewString("Martin"),
			InstrumentName: "Clarinette",
			BirthDate:      nulls.NewTime(bt),
		},
	}
	svc := services.MembershipService{Membership: stub}

	p, err := svc.GetProfile(nil, 1)
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "Alice", p.FirstName)
	assert.NotNil(t, p.BirthDate)
}

func TestGetProfile_Success_NoBirthDate(t *testing.T) {
	stub := &stubMembershipRepo{
		profileRow: &models.MusicianProfileRow{
			ID:             2,
			InstrumentName: "Trompette",
		},
	}
	svc := services.MembershipService{Membership: stub}

	p, err := svc.GetProfile(nil, 2)
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Nil(t, p.BirthDate)
}

func TestGetProfile_Error(t *testing.T) {
	stub := &stubMembershipRepo{profileErr: errDBFailure}
	svc := services.MembershipService{Membership: stub}

	p, err := svc.GetProfile(nil, 1)
	assert.Nil(t, p)
	assert.ErrorIs(t, err, errDBFailure)
}

// --- ToggleProcessingRestriction ---

func TestToggleProcessingRestriction_Success(t *testing.T) {
	stub := &stubMembershipRepo{}
	svc := services.MembershipService{Membership: stub}
	assert.NoError(t, svc.ToggleProcessingRestriction(nil, 1))
}

func TestToggleProcessingRestriction_Error(t *testing.T) {
	stub := &stubMembershipRepo{toggleErr: errDBFailure}
	svc := services.MembershipService{Membership: stub}
	assert.ErrorIs(t, svc.ToggleProcessingRestriction(nil, 1), errDBFailure)
}
