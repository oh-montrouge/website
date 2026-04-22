package services_test

import (
	"errors"
	"testing"

	"github.com/gobuffalo/pop/v6"
	"github.com/stretchr/testify/assert"
	"ohmontrouge/webapp/models"
	"ohmontrouge/webapp/services"
)

// stubSessionRepo is a stub SessionRepository for compliance unit tests.
type stubSessionRepo struct {
	bindErr      error
	deleteErr    error
	deleteCalled bool
}

func (s *stubSessionRepo) BindAccount(_ *pop.Connection, _ string, _ int64) error {
	return s.bindErr
}

func (s *stubSessionRepo) DeleteByAccount(_ *pop.Connection, _ int64) error {
	s.deleteCalled = true
	return s.deleteErr
}

// stubRoleRepoFull is an extended role stub that supports CountActiveAdmins, RevokeRole, RemoveAllRoles.
type stubRoleRepoFull struct {
	hasRole         bool
	hasRoleErr      error
	hasActiveHolder bool
	holderErr       error
	roleID          int64
	roleIDErr       error
	assignErr       error
	adminCount      int
	adminCountErr   error
	revokeErr       error
	removeAllErr    error
	removeAllCalled bool
}

func (s *stubRoleRepoFull) HasRole(_ *pop.Connection, _ int64, _ string) (bool, error) {
	return s.hasRole, s.hasRoleErr
}

func (s *stubRoleRepoFull) HasActiveRoleHolder(_ *pop.Connection, _ string) (bool, error) {
	return s.hasActiveHolder, s.holderErr
}

func (s *stubRoleRepoFull) GetIDByName(_ *pop.Connection, _ string) (int64, error) {
	return s.roleID, s.roleIDErr
}

func (s *stubRoleRepoFull) AssignRole(_ *pop.Connection, _, _ int64) error {
	return s.assignErr
}

func (s *stubRoleRepoFull) CountActiveAdmins(_ *pop.Connection) (int, error) {
	return s.adminCount, s.adminCountErr
}

func (s *stubRoleRepoFull) RevokeRole(_ *pop.Connection, _, _ int64) error {
	return s.revokeErr
}

func (s *stubRoleRepoFull) RemoveAllRoles(_ *pop.Connection, _ int64) error {
	s.removeAllCalled = true
	return s.removeAllErr
}

// stubAccountRepoFull extends stubAccountRepo with AnonymizeAccount tracking.
type stubAccountRepoFull struct {
	account         *models.Account
	err             error
	createdID       int64
	anonymizeCalled bool
	anonymizeErr    error
	deleteCalled    bool
}

func (s *stubAccountRepoFull) FindByEmail(_ *pop.Connection, _ string) (*models.Account, error) {
	return s.account, s.err
}

func (s *stubAccountRepoFull) GetByID(_ *pop.Connection, _ int64) (*models.Account, error) {
	return s.account, s.err
}

func (s *stubAccountRepoFull) Create(_ *pop.Connection, _, _ string, _ int64) (int64, error) {
	return s.createdID, s.err
}

func (s *stubAccountRepoFull) UpdatePasswordHash(_ *pop.Connection, _ int64, _ string) error {
	return s.err
}

func (s *stubAccountRepoFull) Activate(_ *pop.Connection, _ int64, _ string, _ bool) error {
	return s.err
}

func (s *stubAccountRepoFull) CreatePending(_ *pop.Connection, _ string, _ int64) (int64, error) {
	return s.createdID, s.err
}

func (s *stubAccountRepoFull) UpdateEmail(_ *pop.Connection, _ int64, _ string) error {
	return s.err
}

func (s *stubAccountRepoFull) Delete(_ *pop.Connection, _ int64) error {
	s.deleteCalled = true
	return s.err
}

func (s *stubAccountRepoFull) AnonymizeAccount(_ *pop.Connection, _ int64, _ string) error {
	s.anonymizeCalled = true
	return s.anonymizeErr
}

// --- AC-M3: Last-admin protection on RevokeAdmin (tested via AccountService) ---

func TestRevokeAdmin_LastAdmin_ReturnsError(t *testing.T) {
	roles := &stubRoleRepoFull{hasRole: true, adminCount: 1}
	svc := services.AccountService{
		Accounts:     &stubAccountRepoFull{},
		Roles:        roles,
		InviteTokens: &stubInviteTokenRepo{},
		ResetTokens:  &stubResetTokenRepo{},
	}
	err := svc.RevokeAdmin(nil, 42)
	assert.ErrorIs(t, err, services.ErrLastAdmin)
}

func TestRevokeAdmin_NotLastAdmin_Succeeds(t *testing.T) {
	roles := &stubRoleRepoFull{hasRole: true, adminCount: 2, roleID: 1}
	svc := services.AccountService{
		Accounts:     &stubAccountRepoFull{},
		Roles:        roles,
		InviteTokens: &stubInviteTokenRepo{},
		ResetTokens:  &stubResetTokenRepo{},
	}
	err := svc.RevokeAdmin(nil, 42)
	assert.NoError(t, err)
}

func TestRevokeAdmin_NotAdmin_IsIdempotent(t *testing.T) {
	roles := &stubRoleRepoFull{hasRole: false}
	svc := services.AccountService{
		Accounts:     &stubAccountRepoFull{},
		Roles:        roles,
		InviteTokens: &stubInviteTokenRepo{},
		ResetTokens:  &stubResetTokenRepo{},
	}
	err := svc.RevokeAdmin(nil, 42)
	assert.NoError(t, err)
}

// --- AC-M4: Last-admin protection on Anonymize ---

func TestAnonymize_LastAdmin_ReturnsError(t *testing.T) {
	accounts := &stubAccountRepoFull{}
	roles := &stubRoleRepoFull{hasRole: true, adminCount: 1}
	membership := &stubMembershipRepo{}
	invites := &stubInviteTokenRepo{}
	resets := &stubResetTokenRepo{}
	sessions := &stubSessionRepo{}

	svc := services.ComplianceService{
		Accounts:     accounts,
		Membership:   membership,
		Roles:        roles,
		InviteTokens: invites,
		ResetTokens:  resets,
		Sessions:     sessions,
	}

	err := svc.Anonymize(nil, 42)
	assert.ErrorIs(t, err, services.ErrLastAdmin)
	assert.False(t, accounts.anonymizeCalled, "AnonymizeAccount must not be called for last admin")
}

// --- AC-M5: Anonymize correctness — all repos called in correct order ---

func TestAnonymize_NonAdmin_CallsAllReposInOrder(t *testing.T) {
	accounts := &stubAccountRepoFull{}
	roles := &stubRoleRepoFull{hasRole: false}
	membership := &stubMembershipRepo{}
	invites := &stubInviteTokenRepo{}
	resets := &stubResetTokenRepo{}
	sessions := &stubSessionRepo{}

	svc := services.ComplianceService{
		Accounts:     accounts,
		Membership:   membership,
		Roles:        roles,
		InviteTokens: invites,
		ResetTokens:  resets,
		Sessions:     sessions,
	}

	err := svc.Anonymize(nil, 5)
	assert.NoError(t, err)
	assert.True(t, accounts.anonymizeCalled, "AnonymizeAccount should be called")
	assert.True(t, roles.removeAllCalled, "RemoveAllRoles should be called")
	assert.True(t, invites.invalidateCalled, "InviteTokens.InvalidateExisting should be called")
	assert.True(t, resets.invalidateCalled, "ResetTokens.InvalidateExisting should be called")
	assert.True(t, sessions.deleteCalled, "Sessions.DeleteByAccount should be called")
}

func TestAnonymize_MultipleAdmins_Succeeds(t *testing.T) {
	accounts := &stubAccountRepoFull{}
	roles := &stubRoleRepoFull{hasRole: true, adminCount: 2}
	membership := &stubMembershipRepo{}
	invites := &stubInviteTokenRepo{}
	resets := &stubResetTokenRepo{}
	sessions := &stubSessionRepo{}

	svc := services.ComplianceService{
		Accounts:     accounts,
		Membership:   membership,
		Roles:        roles,
		InviteTokens: invites,
		ResetTokens:  resets,
		Sessions:     sessions,
	}

	err := svc.Anonymize(nil, 42)
	assert.NoError(t, err)
	assert.True(t, accounts.anonymizeCalled)
}

func TestAnonymize_AnonymizeAccountError_StopsChain(t *testing.T) {
	accounts := &stubAccountRepoFull{anonymizeErr: errors.New("db error")}
	roles := &stubRoleRepoFull{hasRole: false}
	membership := &stubMembershipRepo{}
	invites := &stubInviteTokenRepo{}
	resets := &stubResetTokenRepo{}
	sessions := &stubSessionRepo{}

	svc := services.ComplianceService{
		Accounts:     accounts,
		Membership:   membership,
		Roles:        roles,
		InviteTokens: invites,
		ResetTokens:  resets,
		Sessions:     sessions,
	}

	err := svc.Anonymize(nil, 5)
	assert.Error(t, err)
	assert.False(t, roles.removeAllCalled, "RemoveAllRoles must not be called if AnonymizeAccount fails")
}

// --- RetentionReviewList ---

func TestRetentionReviewList_MapsRows(t *testing.T) {
	membership := &stubMembershipRepo{
		retentionRows: []models.RetentionRow{
			{ID: 1, FirstName: "Jean", LastName: "Valjean", InstrumentName: "Tuba", LastSeasonLabel: "2016-2017"},
		},
	}
	svc := services.ComplianceService{
		Membership: membership,
	}

	entries, err := svc.RetentionReviewList(nil)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, int64(1), entries[0].AccountID)
	assert.Equal(t, "Jean", entries[0].FirstName)
	assert.Equal(t, "2016-2017", entries[0].LastSeasonLabel)
}
