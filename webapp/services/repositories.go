package services

import (
	"github.com/gobuffalo/pop/v6"
	"ohmontrouge/webapp/models"
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
}

// SessionRepository is the interface the auth handler depends on to link a session to an account.
// The real implementation is models.HTTPSessionStore; tests inject stubs or nil.
type SessionRepository interface {
	BindAccount(db *pop.Connection, sessionKey string, accountID int64) error
}

// RoleRepository is the interface services depend on to check role membership.
// The real implementation is models.AccountRoleStore; tests inject stubs.
type RoleRepository interface {
	HasRole(tx *pop.Connection, accountID int64, roleName string) (bool, error)
	HasActiveRoleHolder(tx *pop.Connection, roleName string) (bool, error)
	GetIDByName(tx *pop.Connection, name string) (int64, error)
	AssignRole(tx *pop.Connection, accountID, roleID int64) error
}

// MembershipRepository is the interface MembershipService and ComplianceService depend on
// to access and mutate musician profile data.
// The real implementation will be added to models.AccountStore in Phase 4.3.
//
// TODO(phase-4.3): add methods as musician management features are implemented.
// Expected: GetProfile, UpdateProfile, ListActive, ListForRetentionReview, ClearProfileFields.
type MembershipRepository any // TODO(phase-4.3): define methods as musician management features are implemented
