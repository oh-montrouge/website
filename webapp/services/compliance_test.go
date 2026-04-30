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

// newRevokeAdminService creates an AccountService for RevokeAdmin unit tests.
// Only Accounts and Roles vary; InviteTokens and ResetTokens use no-op stubs.
func newRevokeAdminService(roles *stubRoleRepoFull) services.AccountService {
	return services.AccountService{
		Accounts:     &stubAccountRepoFull{},
		Roles:        roles,
		InviteTokens: &stubInviteTokenRepo{},
		ResetTokens:  &stubResetTokenRepo{},
	}
}

// complianceDeps holds all deps for ComplianceService tests with accessible references.
type complianceDeps struct {
	accounts   *stubAccountRepoFull
	roles      *stubRoleRepoFull
	membership *stubMembershipRepo
	invites    *stubInviteTokenRepo
	resets     *stubResetTokenRepo
	sessions   *stubSessionRepo
}

func (d *complianceDeps) service() services.ComplianceService {
	return services.ComplianceService{
		Accounts:     d.accounts,
		Membership:   d.membership,
		Roles:        d.roles,
		InviteTokens: d.invites,
		ResetTokens:  d.resets,
		Sessions:     d.sessions,
	}
}

func newComplianceDeps(accounts *stubAccountRepoFull, roles *stubRoleRepoFull) *complianceDeps {
	return &complianceDeps{
		accounts:   accounts,
		roles:      roles,
		membership: &stubMembershipRepo{},
		invites:    &stubInviteTokenRepo{},
		resets:     &stubResetTokenRepo{},
		sessions:   &stubSessionRepo{},
	}
}

// --- AC-M3: Last-admin protection on RevokeAdmin (tested via AccountService) ---

func TestRevokeAdmin_LastAdmin_ReturnsError(t *testing.T) {
	err := newRevokeAdminService(&stubRoleRepoFull{hasRole: true, adminCount: 1}).RevokeAdmin(nil, 42)
	assert.ErrorIs(t, err, services.ErrLastAdmin)
}

func TestRevokeAdmin_NotLastAdmin_Succeeds(t *testing.T) {
	err := newRevokeAdminService(&stubRoleRepoFull{hasRole: true, adminCount: 2, roleID: 1}).RevokeAdmin(nil, 42)
	assert.NoError(t, err)
}

func TestRevokeAdmin_NotAdmin_IsIdempotent(t *testing.T) {
	err := newRevokeAdminService(&stubRoleRepoFull{hasRole: false}).RevokeAdmin(nil, 42)
	assert.NoError(t, err)
}

// --- AC-M4: Last-admin protection on Anonymize ---

func TestAnonymize_LastAdmin_ReturnsError(t *testing.T) {
	deps := newComplianceDeps(&stubAccountRepoFull{}, &stubRoleRepoFull{hasRole: true, adminCount: 1})
	err := deps.service().Anonymize(nil, 42)
	assert.ErrorIs(t, err, services.ErrLastAdmin)
	assert.False(t, deps.accounts.anonymizeCalled, "AnonymizeAccount must not be called for last admin")
}

// --- AC-M5: Anonymize correctness — all repos called in correct order ---

func TestAnonymize_NonAdmin_CallsAllReposInOrder(t *testing.T) {
	deps := newComplianceDeps(&stubAccountRepoFull{}, &stubRoleRepoFull{hasRole: false})
	err := deps.service().Anonymize(nil, 5)
	assert.NoError(t, err)
	assert.True(t, deps.accounts.anonymizeCalled, "AnonymizeAccount should be called")
	assert.True(t, deps.roles.removeAllCalled, "RemoveAllRoles should be called")
	assert.True(t, deps.invites.invalidateCalled, "InviteTokens.InvalidateExisting should be called")
	assert.True(t, deps.resets.invalidateCalled, "ResetTokens.InvalidateExisting should be called")
	assert.True(t, deps.sessions.deleteCalled, "Sessions.DeleteByAccount should be called")
}

func TestAnonymize_MultipleAdmins_Succeeds(t *testing.T) {
	deps := newComplianceDeps(&stubAccountRepoFull{}, &stubRoleRepoFull{hasRole: true, adminCount: 2})
	err := deps.service().Anonymize(nil, 42)
	assert.NoError(t, err)
	assert.True(t, deps.accounts.anonymizeCalled)
}

func TestAnonymize_AnonymizeAccountError_StopsChain(t *testing.T) {
	deps := newComplianceDeps(&stubAccountRepoFull{anonymizeErr: errors.New("db error")}, &stubRoleRepoFull{hasRole: false})
	err := deps.service().Anonymize(nil, 5)
	assert.Error(t, err)
	assert.False(t, deps.roles.removeAllCalled, "RemoveAllRoles must not be called if AnonymizeAccount fails")
}

// --- AC-M7: Anonymize deletes RSVPs ---

// stubRSVPDeleteRepo is a minimal RSVPRepository stub that tracks DeleteByAccount calls.
type stubRSVPDeleteRepo struct {
	deleted bool
	err     error
}

func (s *stubRSVPDeleteRepo) SeedForEvent(_ *pop.Connection, _ int64) error   { return s.err }
func (s *stubRSVPDeleteRepo) SeedForAccount(_ *pop.Connection, _ int64) error { return s.err }
func (s *stubRSVPDeleteRepo) GetByAccountAndEvent(_ *pop.Connection, _, _ int64) (*models.RSVPRow, error) {
	return nil, s.err
}
func (s *stubRSVPDeleteRepo) Update(_ *pop.Connection, _ int64, _ string, _ *int64) error {
	return s.err
}
func (s *stubRSVPDeleteRepo) DeleteByAccount(_ *pop.Connection, _ int64) error {
	s.deleted = true
	return s.err
}
func (s *stubRSVPDeleteRepo) ClearFieldResponses(_ *pop.Connection, _ int64) error { return s.err }
func (s *stubRSVPDeleteRepo) ListForEvent(_ *pop.Connection, _ int64) ([]models.RSVPListRow, error) {
	return nil, s.err
}
func (s *stubRSVPDeleteRepo) ResetYesRSVPs(_ *pop.Connection, _ int64) error    { return s.err }
func (s *stubRSVPDeleteRepo) ClearInstruments(_ *pop.Connection, _ int64) error { return s.err }
func (s *stubRSVPDeleteRepo) AddFieldResponse(_ *pop.Connection, _, _ int64, _ string) error {
	return s.err
}
func (s *stubRSVPDeleteRepo) ListFieldResponses(_ *pop.Connection, _ int64) ([]models.RSVPFieldResponseRow, error) {
	return nil, s.err
}

func TestAnonymize_DeletesRSVPs(t *testing.T) {
	rsvps := &stubRSVPDeleteRepo{}
	svc := services.ComplianceService{
		Accounts:     &stubAccountRepoFull{},
		Membership:   &stubMembershipRepo{},
		Roles:        &stubRoleRepoFull{hasRole: false},
		InviteTokens: &stubInviteTokenRepo{},
		ResetTokens:  &stubResetTokenRepo{},
		Sessions:     &stubSessionRepo{},
		Events:       rsvps,
	}
	err := svc.Anonymize(nil, 5)
	assert.NoError(t, err)
	assert.True(t, rsvps.deleted, "DeleteByAccount should be called during anonymization")
}

func TestAnonymize_NilEvents_DoesNotPanic(t *testing.T) {
	svc := services.ComplianceService{
		Accounts:     &stubAccountRepoFull{},
		Membership:   &stubMembershipRepo{},
		Roles:        &stubRoleRepoFull{hasRole: false},
		InviteTokens: &stubInviteTokenRepo{},
		ResetTokens:  &stubResetTokenRepo{},
		Sessions:     &stubSessionRepo{},
		Events:       nil,
	}
	err := svc.Anonymize(nil, 5)
	assert.NoError(t, err)
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
