package services

import (
	"errors"
	"time"

	"github.com/gobuffalo/pop/v6"
	"ohmontrouge/webapp/models"
)

var ErrParentalConsentRequired = errors.New("l'accord parental est obligatoire pour les musiciens de moins de 15 ans")

// MusicianProfile is the Membership-context view of an account.
// It carries all musician profile and GDPR-sensitive fields owned by the Membership bounded
// context. Auth fields (email, password, status) are absent — those belong to AccountDTO.
// See specs/technical-adrs/007-account-musician-dtos.md and architecture/context-map.md § Membership.
type MusicianProfile struct {
	AccountID            int64
	FirstName            string
	LastName             string
	MainInstrumentID     int64
	MainInstrumentName   string
	BirthDate            *time.Time
	ParentalConsentURI   string
	Phone                string
	Address              string
	PhoneAddressConsent  bool
	ProcessingRestricted bool
}

// MusicianProfileSummary is used for admin list pages.
type MusicianProfileSummary struct {
	AccountID          int64
	FirstName          string
	LastName           string
	AnonymizationToken string
	MainInstrumentName string
	Status             string
	IsAdmin            bool
}

// RetentionEntryDTO carries the data for the retention review list.
type RetentionEntryDTO struct {
	AccountID          int64
	FirstName          string
	LastName           string
	MainInstrumentName string
	LastSeasonLabel    string
	LastSeasonEndDate  time.Time
}

// MusicianProfileManager is the interface handlers depend on for musician profile operations.
type MusicianProfileManager interface {
	GetProfile(tx *pop.Connection, accountID int64) (*MusicianProfile, error)
	SetInitialProfile(tx *pop.Connection, accountID int64, firstName, lastName string, birthDate *time.Time, parentalConsentURI string) error
	UpdateProfile(tx *pop.Connection, accountID int64, firstName, lastName, email string, instrumentID int64, birthDate *time.Time, parentalConsentURI, phone, address string) error
	ListNonAnonymized(tx *pop.Connection) ([]MusicianProfileSummary, error)
	ConsentWithdrawal(tx *pop.Connection, accountID int64) error
	ToggleProcessingRestriction(tx *pop.Connection, accountID int64) error
}

// MembershipService owns all musician profile operations.
// It is the sole writer of Membership-context fields on the accounts table.
// Auth fields are written exclusively by AccountService.
// Cross-context operations (anonymization) are coordinated by ComplianceService.
type MembershipService struct {
	Membership MembershipRepository
}

func (s MembershipService) GetProfile(tx *pop.Connection, accountID int64) (*MusicianProfile, error) {
	row, err := s.Membership.GetProfile(tx, accountID)
	if err != nil {
		return nil, err
	}
	return toMusicianProfile(row), nil
}

func (s MembershipService) SetInitialProfile(tx *pop.Connection, accountID int64, firstName, lastName string, birthDate *time.Time, parentalConsentURI string) error {
	if err := validateUnder15(birthDate, parentalConsentURI); err != nil {
		return err
	}
	return s.Membership.SetProfile(tx, accountID, firstName, lastName, birthDate, parentalConsentURI)
}

func (s MembershipService) UpdateProfile(tx *pop.Connection, accountID int64, firstName, lastName, email string, instrumentID int64, birthDate *time.Time, parentalConsentURI, phone, address string) error {
	if err := validateUnder15(birthDate, parentalConsentURI); err != nil {
		return err
	}
	return s.Membership.UpdateProfile(tx, accountID, firstName, lastName, email, instrumentID, birthDate, parentalConsentURI, phone, address)
}

func (s MembershipService) ListNonAnonymized(tx *pop.Connection) ([]MusicianProfileSummary, error) {
	rows, err := s.Membership.ListNonAnonymized(tx)
	if err != nil {
		return nil, err
	}
	summaries := make([]MusicianProfileSummary, len(rows))
	for i, r := range rows {
		summaries[i] = MusicianProfileSummary{
			AccountID:          r.ID,
			FirstName:          r.FirstName.String,
			LastName:           r.LastName.String,
			AnonymizationToken: r.AnonymizationToken.String,
			MainInstrumentName: r.InstrumentName,
			Status:             r.Status,
			IsAdmin:            r.IsAdmin,
		}
	}
	return summaries, nil
}

func (s MembershipService) ConsentWithdrawal(tx *pop.Connection, accountID int64) error {
	return s.Membership.WithdrawConsent(tx, accountID)
}

func (s MembershipService) ToggleProcessingRestriction(tx *pop.Connection, accountID int64) error {
	return s.Membership.ToggleProcessingRestriction(tx, accountID)
}

func validateUnder15(birthDate *time.Time, parentalConsentURI string) error {
	if birthDate == nil {
		return nil
	}
	if isUnder15(*birthDate) && parentalConsentURI == "" {
		return ErrParentalConsentRequired
	}
	return nil
}

func isUnder15(birthDate time.Time) bool {
	return birthDate.After(time.Now().AddDate(-15, 0, 0))
}

func toMusicianProfile(row *models.MusicianProfileRow) *MusicianProfile {
	p := &MusicianProfile{
		AccountID:            row.ID,
		FirstName:            row.FirstName.String,
		LastName:             row.LastName.String,
		MainInstrumentID:     row.MainInstrumentID,
		MainInstrumentName:   row.InstrumentName,
		ParentalConsentURI:   row.ParentalConsentURI.String,
		Phone:                row.Phone.String,
		Address:              row.Address.String,
		PhoneAddressConsent:  row.PhoneAddressConsent,
		ProcessingRestricted: row.ProcessingRestricted,
	}
	if row.BirthDate.Valid {
		t := row.BirthDate.Time
		p.BirthDate = &t
	}
	return p
}
