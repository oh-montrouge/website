package services

// ComplianceService coordinates GDPR operations that span multiple bounded contexts.
// It is the authority on anonymization, consent withdrawal, and processing restriction —
// operations the context map assigns to the Compliance context.
//
// Ownership rule: neither AccountService nor MembershipService reaches into the other.
// ComplianceService is the only place where both contexts are touched in a single transaction.
// See specs/technical-adrs/007-account-musician-dtos.md and architecture/context-map.md § Compliance.
//
// TODO(phase-4.3): implement Anonymize, ToggleProcessingRestriction.
// Anonymize must clear I&A fields via AccountRepository and Membership fields via
// MembershipRepository atomically. It must also delete all RSVP records (Event Coordination
// context) — wire in an EventRepository when Phase 4.6 is implemented.
type ComplianceService struct {
	Accounts   AccountRepository
	Membership MembershipRepository
	Roles      RoleRepository
}
