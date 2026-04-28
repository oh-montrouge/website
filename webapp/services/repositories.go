package services

import (
	"time"

	"ohmontrouge/webapp/models"

	"github.com/gobuffalo/pop/v6"
)

// InstrumentRepository is the interface services and handlers depend on to access instrument data.
// Defined here (consumer side) per the Dependency Inversion Principle.
// The real implementation is models.InstrumentStore; tests inject stubs.
type InstrumentRepository interface {
	List(tx *pop.Connection) (models.Instruments, error)
}

// AccountRepository is the interface services depend on to access account data.
// The real implementation is models.AccountStore; tests inject stubs.
type AccountRepository interface {
	FindByEmail(tx *pop.Connection, email string) (*models.Account, error)
	GetByID(tx *pop.Connection, id int64) (*models.Account, error)
	Create(tx *pop.Connection, email, passwordHash string, instrumentID int64) (int64, error)
	UpdatePasswordHash(tx *pop.Connection, id int64, hash string) error
	// Activate transitions a pending account to active in a single SQL UPDATE.
	// Sets status, password_hash, phone_address_consent; clears phone/address when consent is false.
	Activate(tx *pop.Connection, id int64, passwordHash string, phoneAddressConsent bool) error
	CreatePending(tx *pop.Connection, email string, instrumentID int64) (int64, error)
	UpdateEmail(tx *pop.Connection, id int64, email string) error
	Delete(tx *pop.Connection, id int64) error
	AnonymizeAccount(tx *pop.Connection, id int64, token string) error
}

// InviteTokenRepository is the interface for invite token persistence.
// The real implementation is models.InviteTokenStore; tests inject stubs.
type InviteTokenRepository interface {
	// Generate inserts a new token row. Call InvalidateExisting first within the same tx.
	Generate(tx *pop.Connection, accountID int64, token string, expiresAt time.Time) error
	// FindByToken returns the token row joined with account and instrument data.
	// Returns nil, nil when the token is not found, expired, used, or the account is not pending.
	FindByToken(tx *pop.Connection, token string) (*models.InviteTokenRecord, error)
	// MarkUsed marks the given token as used.
	MarkUsed(tx *pop.Connection, tokenID int64) error
	// InvalidateExisting marks all unused tokens for the account as used.
	InvalidateExisting(tx *pop.Connection, accountID int64) error
	// FindActiveForAccount returns the current active (unused, non-expired) invite token, or nil.
	FindActiveForAccount(tx *pop.Connection, accountID int64) (*models.InviteToken, error)
}

// PasswordResetTokenRepository is the interface for password reset token persistence.
// The real implementation is models.PasswordResetTokenStore; tests inject stubs.
type PasswordResetTokenRepository interface {
	// Generate inserts a new token row. Call InvalidateExisting first within the same tx.
	Generate(tx *pop.Connection, accountID int64, token string, expiresAt time.Time) error
	// FindByToken returns the token row joined with account status.
	// Returns nil, nil when the token is not found, expired, used, or the account is not active.
	FindByToken(tx *pop.Connection, token string) (*models.PasswordResetTokenRecord, error)
	// MarkUsed marks the given token as used.
	MarkUsed(tx *pop.Connection, tokenID int64) error
	// InvalidateExisting marks all unused tokens for the account as used.
	InvalidateExisting(tx *pop.Connection, accountID int64) error
	// FindActiveForAccount returns the current active (unused, non-expired) reset token, or nil.
	FindActiveForAccount(tx *pop.Connection, accountID int64) (*models.PasswordResetToken, error)
}

// SessionRepository is the interface the auth handler depends on to link a session to an account.
// The real implementation is models.HTTPSessionStore; tests inject stubs or nil.
type SessionRepository interface {
	BindAccount(db *pop.Connection, sessionKey string, accountID int64) error
	DeleteByAccount(db *pop.Connection, accountID int64) error
}

// RoleRepository is the interface services depend on to check role membership.
// The real implementation is models.AccountRoleStore; tests inject stubs.
type RoleRepository interface {
	HasRole(tx *pop.Connection, accountID int64, roleName string) (bool, error)
	HasActiveRoleHolder(tx *pop.Connection, roleName string) (bool, error)
	GetIDByName(tx *pop.Connection, name string) (int64, error)
	AssignRole(tx *pop.Connection, accountID, roleID int64) error
	CountActiveAdmins(tx *pop.Connection) (int, error)
	RevokeRole(tx *pop.Connection, accountID, roleID int64) error
	RemoveAllRoles(tx *pop.Connection, accountID int64) error
}

// MembershipRepository is the interface MembershipService and ComplianceService depend on
// to access and mutate musician profile data. Implemented by models.AccountStore.
type MembershipRepository interface {
	GetProfile(tx *pop.Connection, accountID int64) (*models.MusicianProfileRow, error)
	SetProfile(tx *pop.Connection, accountID int64, firstName, lastName string, birthDate *time.Time, parentalConsentURI string) error
	// UpdateProfile includes email and instrumentID despite those being I&A fields on the
	// accounts table. Keeping them here avoids handler-level composition for the single-intent
	// "edit profile" form. AccountRepository.UpdateEmail exists but is intentionally unused
	// for this path.
	UpdateProfile(tx *pop.Connection, accountID int64, firstName, lastName, email string, instrumentID int64, birthDate *time.Time, parentalConsentURI, phone, address string) error
	ListNonAnonymized(tx *pop.Connection) ([]models.MusicianListRow, error)
	ListForRetentionReview(tx *pop.Connection) ([]models.RetentionRow, error)
	ClearMembershipFields(tx *pop.Connection, accountID int64) error
	WithdrawConsent(tx *pop.Connection, accountID int64) error
	ToggleProcessingRestriction(tx *pop.Connection, accountID int64) error
}

// SeasonRepository is the interface SeasonService depends on to access season data.
// The real implementation is models.SeasonStore; tests inject stubs.
type SeasonRepository interface {
	Create(tx *pop.Connection, label string, startDate, endDate time.Time) (int64, error)
	List(tx *pop.Connection) (models.Seasons, error)
	DesignateCurrent(tx *pop.Connection, id int64) error
}

// FeePaymentRepository is the interface FeePaymentService depends on to access fee payment data.
// The real implementation is models.FeePaymentStore; tests inject stubs.
type FeePaymentRepository interface {
	Create(tx *pop.Connection, accountID, seasonID int64, amount float64, paymentDate time.Time, paymentType, comment string) (int64, error)
	Update(tx *pop.Connection, id int64, amount float64, paymentDate time.Time, paymentType, comment string) error
	Delete(tx *pop.Connection, id int64) error
	ListByAccount(tx *pop.Connection, accountID int64) (models.FeePaymentRows, error)
	GetByID(tx *pop.Connection, id int64) (*models.FeePaymentRow, error)
	GetFirstInscriptionDate(tx *pop.Connection, accountID int64) (*time.Time, error)
}

// EventRepository is the interface EventService depends on to access event and field data.
// The real implementation is models.EventStore; tests inject stubs.
type EventRepository interface {
	Create(tx *pop.Connection, name, eventType string, datetime time.Time) (int64, error)
	GetByID(tx *pop.Connection, id int64) (*models.EventDetailRow, error)
	Update(tx *pop.Connection, id int64, name, eventType string, datetime time.Time) error
	Delete(tx *pop.Connection, id int64) error
	ListUpcoming(tx *pop.Connection, accountID int64) ([]models.EventListRow, error)
	ListAll(tx *pop.Connection, accountID int64) ([]models.EventListRow, error)
	DeleteFields(tx *pop.Connection, eventID int64) error
	AddField(tx *pop.Connection, eventID int64, label, fieldType string, required bool, position int) (int64, error)
	GetFieldByID(tx *pop.Connection, fieldID int64) (*models.EventFieldRow, error)
	UpdateField(tx *pop.Connection, fieldID int64, label, fieldType string, required bool, position int) error
	DeleteField(tx *pop.Connection, fieldID int64) error
	ListFields(tx *pop.Connection, eventID int64) ([]models.EventFieldRow, error)
	ListFieldChoices(tx *pop.Connection, fieldID int64) ([]models.EventFieldChoiceRow, error)
	CountFieldResponses(tx *pop.Connection, fieldID int64) (int, error)
	AddFieldChoice(tx *pop.Connection, fieldID int64, label string, position int) (int64, error)
	DeleteFieldChoices(tx *pop.Connection, fieldID int64) error
}

// RSVPRepository is the interface EventService depends on to access RSVP data.
// The real implementation is models.RSVPStore; tests inject stubs.
type RSVPRepository interface {
	SeedForEvent(tx *pop.Connection, eventID int64) error
	SeedForAccount(tx *pop.Connection, accountID int64) error
	GetByAccountAndEvent(tx *pop.Connection, accountID, eventID int64) (*models.RSVPRow, error)
	Update(tx *pop.Connection, rsvpID int64, state string, instrumentID *int64) error
	DeleteByAccount(tx *pop.Connection, accountID int64) error
	ClearFieldResponses(tx *pop.Connection, rsvpID int64) error
	ListForEvent(tx *pop.Connection, eventID int64) ([]models.RSVPListRow, error)
	ResetYesRSVPs(tx *pop.Connection, eventID int64) error
	ClearInstruments(tx *pop.Connection, eventID int64) error
	AddFieldResponse(tx *pop.Connection, rsvpID, fieldID int64, value string) error
	ListFieldResponses(tx *pop.Connection, eventID int64) ([]models.RSVPFieldResponseRow, error)
}
