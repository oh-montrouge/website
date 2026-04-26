package services

import (
	"github.com/gobuffalo/pop/v6"
)

// ComplianceManager is the interface handlers depend on for compliance operations.
type ComplianceManager interface {
	Anonymize(tx *pop.Connection, accountID int64) error
	RetentionReviewList(tx *pop.Connection) ([]RetentionEntryDTO, error)
}

// ComplianceService coordinates GDPR operations that span multiple bounded contexts.
// It is the authority on anonymization, consent withdrawal, and processing restriction —
// operations the context map assigns to the Compliance context.
//
// Ownership rule: neither AccountService nor MembershipService reaches into the other.
// ComplianceService is the only place where both contexts are touched in a single transaction.
// See specs/technical-adrs/007-account-musician-dtos.md and architecture/context-map.md § Compliance.
type ComplianceService struct {
	Accounts     AccountRepository
	Membership   MembershipRepository
	Roles        RoleRepository
	InviteTokens InviteTokenRepository
	ResetTokens  PasswordResetTokenRepository
	Sessions     SessionRepository
	Events       RSVPRepository // for RSVP deletion on anonymization (Phase 4.6)
}

// Anonymize performs an atomic GDPR erasure on the given account.
// Last-admin protection is checked before any mutation.
func (s ComplianceService) Anonymize(tx *pop.Connection, accountID int64) error {
	isAdmin, err := s.Roles.HasRole(tx, accountID, RoleAdmin)
	if err != nil {
		return err
	}
	if isAdmin {
		count, err := s.Roles.CountActiveAdmins(tx)
		if err != nil {
			return err
		}
		if count <= 1 {
			return ErrLastAdmin
		}
	}

	token, err := generateCSPRNGToken()
	if err != nil {
		return err
	}

	if err := s.Accounts.AnonymizeAccount(tx, accountID, token); err != nil {
		return err
	}
	if err := s.Membership.ClearMembershipFields(tx, accountID); err != nil {
		return err
	}
	if err := s.Roles.RemoveAllRoles(tx, accountID); err != nil {
		return err
	}
	if s.Events != nil {
		if err := s.Events.DeleteByAccount(tx, accountID); err != nil {
			return err
		}
	}
	if err := s.InviteTokens.InvalidateExisting(tx, accountID); err != nil {
		return err
	}
	if err := s.ResetTokens.InvalidateExisting(tx, accountID); err != nil {
		return err
	}
	return s.Sessions.DeleteByAccount(tx, accountID)
}

// RetentionReviewList returns accounts whose data retention period has elapsed.
func (s ComplianceService) RetentionReviewList(tx *pop.Connection) ([]RetentionEntryDTO, error) {
	rows, err := s.Membership.ListForRetentionReview(tx)
	if err != nil {
		return nil, err
	}
	entries := make([]RetentionEntryDTO, len(rows))
	for i, r := range rows {
		entries[i] = RetentionEntryDTO{
			AccountID:          r.ID,
			FirstName:          r.FirstName,
			LastName:           r.LastName,
			MainInstrumentName: r.InstrumentName,
			LastSeasonLabel:    r.LastSeasonLabel,
			LastSeasonEndDate:  r.LastSeasonEndDate,
		}
	}
	return entries, nil
}
