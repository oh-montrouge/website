package services

import "time"

// MusicianProfile is the Membership-context view of an account.
// It carries all musician profile and GDPR-sensitive fields owned by the Membership bounded
// context. Auth fields (email, password, status) are absent — those belong to AccountDTO.
// See specs/technical-adrs/007-account-musician-dtos.md and architecture/context-map.md § Membership.
type MusicianProfile struct {
	AccountID            int64
	FirstName            string
	LastName             string
	MainInstrumentID     int64
	BirthDate            *time.Time
	ParentalConsentURI   string
	Phone                string
	Address              string
	PhoneAddressConsent  bool
	ProcessingRestricted bool
}

// MusicianProfiles is a slice of MusicianProfile.
type MusicianProfiles []MusicianProfile

// MembershipService owns all musician profile operations.
// It is the sole writer of Membership-context fields on the accounts table.
// Auth fields are written exclusively by AccountService.
// Cross-context operations (anonymization) are coordinated by ComplianceService.
//
// TODO(phase-4.3): implement GetProfile, UpdateProfile, ListActive, ConsentWithdrawal.
type MembershipService struct {
	Membership MembershipRepository
}
