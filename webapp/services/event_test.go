package services_test

import (
	"errors"
	"testing"
	"time"

	"ohmontrouge/webapp/models"
	"ohmontrouge/webapp/services"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/stretchr/testify/assert"
)

// --- EventRepository stub ---

type stubEventRepoE struct {
	event     *models.EventDetailRow
	eventID   int64
	err       error
	fieldID   int64
	fieldErr  error
	field     *models.EventFieldRow
	countResp int
	deleted   bool
	updated   bool
}

func (s *stubEventRepoE) Create(_ *pop.Connection, _, _ string, _ time.Time) (int64, error) {
	return s.eventID, s.err
}
func (s *stubEventRepoE) GetByID(_ *pop.Connection, _ int64) (*models.EventDetailRow, error) {
	return s.event, s.err
}
func (s *stubEventRepoE) Update(_ *pop.Connection, _ int64, _, _ string, _ time.Time) error {
	s.updated = true
	return s.err
}
func (s *stubEventRepoE) Delete(_ *pop.Connection, _ int64) error { return s.err }
func (s *stubEventRepoE) ListUpcoming(_ *pop.Connection, _ int64) ([]models.EventListRow, error) {
	return nil, s.err
}
func (s *stubEventRepoE) ListAll(_ *pop.Connection, _ int64) ([]models.EventListRow, error) {
	return nil, s.err
}
func (s *stubEventRepoE) DeleteFields(_ *pop.Connection, _ int64) error {
	s.deleted = true
	return s.err
}
func (s *stubEventRepoE) AddField(_ *pop.Connection, _ int64, _, _ string, _ bool, _ int) (int64, error) {
	return s.fieldID, s.fieldErr
}
func (s *stubEventRepoE) GetFieldByID(_ *pop.Connection, _ int64) (*models.EventFieldRow, error) {
	return s.field, s.fieldErr
}
func (s *stubEventRepoE) UpdateField(_ *pop.Connection, _ int64, _, _ string, _ bool, _ int) error {
	return s.fieldErr
}
func (s *stubEventRepoE) DeleteField(_ *pop.Connection, _ int64) error { return s.fieldErr }
func (s *stubEventRepoE) ListFields(_ *pop.Connection, _ int64) ([]models.EventFieldRow, error) {
	return nil, s.err
}
func (s *stubEventRepoE) ListFieldChoices(_ *pop.Connection, _ int64) ([]models.EventFieldChoiceRow, error) {
	return nil, s.err
}
func (s *stubEventRepoE) CountFieldResponses(_ *pop.Connection, _ int64) (int, error) {
	return s.countResp, s.fieldErr
}
func (s *stubEventRepoE) AddFieldChoice(_ *pop.Connection, _ int64, _ string, _ int) (int64, error) {
	return 0, s.fieldErr
}
func (s *stubEventRepoE) DeleteFieldChoices(_ *pop.Connection, _ int64) error { return s.fieldErr }

// --- RSVPRepository stub ---

type stubRSVPRepoE struct {
	rsvp            *models.RSVPRow
	err             error
	seeded          bool
	cleared         bool
	resetYes        bool
	clearInstrument bool
}

func (s *stubRSVPRepoE) SeedForEvent(_ *pop.Connection, _ int64) error {
	s.seeded = true
	return s.err
}
func (s *stubRSVPRepoE) SeedForAccount(_ *pop.Connection, _ int64) error {
	s.seeded = true
	return s.err
}
func (s *stubRSVPRepoE) GetByAccountAndEvent(_ *pop.Connection, _, _ int64) (*models.RSVPRow, error) {
	return s.rsvp, s.err
}
func (s *stubRSVPRepoE) Update(_ *pop.Connection, _ int64, _ string, _ *int64) error { return s.err }
func (s *stubRSVPRepoE) DeleteByAccount(_ *pop.Connection, _ int64) error            { return s.err }
func (s *stubRSVPRepoE) ClearFieldResponses(_ *pop.Connection, _ int64) error {
	s.cleared = true
	return s.err
}
func (s *stubRSVPRepoE) ListForEvent(_ *pop.Connection, _ int64) ([]models.RSVPListRow, error) {
	return nil, s.err
}
func (s *stubRSVPRepoE) ResetYesRSVPs(_ *pop.Connection, _ int64) error {
	s.resetYes = true
	return s.err
}
func (s *stubRSVPRepoE) ClearInstruments(_ *pop.Connection, _ int64) error {
	s.clearInstrument = true
	return s.err
}
func (s *stubRSVPRepoE) AddFieldResponse(_ *pop.Connection, _, _ int64, _ string) error {
	return s.err
}
func (s *stubRSVPRepoE) ListFieldResponses(_ *pop.Connection, _ int64) ([]models.RSVPFieldResponseRow, error) {
	return nil, s.err
}

// --- AC-M1: Create seeds RSVPs for all active accounts ---

func TestEventService_Create_SeedsRSVPs(t *testing.T) {
	rsvps := &stubRSVPRepoE{}
	svc := services.EventService{
		Events: &stubEventRepoE{eventID: 42},
		RSVPs:  rsvps,
	}
	err := svc.Create(nil, "Concert de printemps", "concert", time.Now())
	assert.NoError(t, err)
	assert.True(t, rsvps.seeded, "SeedForEvent should have been called")
}

func TestEventService_Create_PropagatesRepoError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := services.EventService{
		Events: &stubEventRepoE{err: repoErr},
		RSVPs:  &stubRSVPRepoE{},
	}
	err := svc.Create(nil, "Concert", "concert", time.Now())
	assert.ErrorIs(t, err, repoErr)
}

// --- AC-M2: Type change rehearsal → concert resets yes RSVPs ---

func TestEventService_Update_RehearsalToConcert_ResetsYesRSVPs(t *testing.T) {
	rsvps := &stubRSVPRepoE{}
	events := &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "rehearsal"}}
	svc := services.EventService{Events: events, RSVPs: rsvps}

	err := svc.Update(nil, 1, "Event", "concert", time.Now())
	assert.NoError(t, err)
	assert.True(t, rsvps.resetYes, "ResetYesRSVPs should have been called")
	assert.False(t, events.deleted, "DeleteFields should not be called")
}

// --- AC-M3: Type change other → concert deletes fields and resets yes RSVPs ---

func TestEventService_Update_OtherToConcert_DeletesFieldsAndResetsYes(t *testing.T) {
	rsvps := &stubRSVPRepoE{}
	events := &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "other"}}
	svc := services.EventService{Events: events, RSVPs: rsvps}

	err := svc.Update(nil, 1, "Event", "concert", time.Now())
	assert.NoError(t, err)
	assert.True(t, events.deleted, "DeleteFields should have been called")
	assert.True(t, rsvps.resetYes, "ResetYesRSVPs should have been called")
}

func TestEventService_Update_OtherToRehearsal_DeletesFields(t *testing.T) {
	rsvps := &stubRSVPRepoE{}
	events := &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "other"}}
	svc := services.EventService{Events: events, RSVPs: rsvps}

	err := svc.Update(nil, 1, "Event", "rehearsal", time.Now())
	assert.NoError(t, err)
	assert.True(t, events.deleted, "DeleteFields should have been called")
	assert.False(t, rsvps.resetYes, "ResetYesRSVPs should not be called")
}

func TestEventService_Update_ConcertToRehearsal_ClearsInstruments(t *testing.T) {
	rsvps := &stubRSVPRepoE{}
	events := &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "concert"}}
	svc := services.EventService{Events: events, RSVPs: rsvps}

	err := svc.Update(nil, 1, "Event", "rehearsal", time.Now())
	assert.NoError(t, err)
	assert.True(t, rsvps.clearInstrument, "ClearInstruments should have been called")
	assert.False(t, rsvps.resetYes)
}

func TestEventService_Update_SameType_NoSideEffects(t *testing.T) {
	rsvps := &stubRSVPRepoE{}
	events := &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "rehearsal"}}
	svc := services.EventService{Events: events, RSVPs: rsvps}

	err := svc.Update(nil, 1, "Event", "rehearsal", time.Now())
	assert.NoError(t, err)
	assert.False(t, rsvps.resetYes)
	assert.False(t, rsvps.clearInstrument)
	assert.False(t, events.deleted)
}

func TestEventService_Update_NotFound(t *testing.T) {
	svc := services.EventService{
		Events: &stubEventRepoE{event: nil},
		RSVPs:  &stubRSVPRepoE{},
	}
	err := svc.Update(nil, 99, "X", "concert", time.Now())
	assert.ErrorIs(t, err, services.ErrEventNotFound)
}

// --- AC-M4: RSVP state change from yes clears field responses ---

func TestEventService_UpdateRSVP_YesToNo_ClearsFieldResponses(t *testing.T) {
	rsvps := &stubRSVPRepoE{rsvp: &models.RSVPRow{ID: 10, State: "yes"}}
	svc := services.EventService{
		Events: &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "rehearsal"}},
		RSVPs:  rsvps,
	}
	err := svc.UpdateRSVP(nil, 1, 2, "no", nil, nil)
	assert.NoError(t, err)
	assert.True(t, rsvps.cleared, "ClearFieldResponses should have been called")
}

func TestEventService_UpdateRSVP_YesToYes_ClearsBeforeResaving(t *testing.T) {
	rsvps := &stubRSVPRepoE{rsvp: &models.RSVPRow{ID: 10, State: "yes"}}
	svc := services.EventService{
		Events: &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "rehearsal"}},
		RSVPs:  rsvps,
	}
	err := svc.UpdateRSVP(nil, 1, 2, "yes", nil, nil)
	assert.NoError(t, err)
	assert.True(t, rsvps.cleared, "ClearFieldResponses must be called so optional fields can be removed")
}

func TestEventService_UpdateRSVP_Concert_YesToNo_ClearsFieldResponses(t *testing.T) {
	instrID := int64(3)
	rsvps := &stubRSVPRepoE{rsvp: &models.RSVPRow{
		ID:           20,
		State:        "yes",
		InstrumentID: nulls.NewInt64(instrID),
	}}
	svc := services.EventService{
		Events: &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "concert"}},
		RSVPs:  rsvps,
	}
	err := svc.UpdateRSVP(nil, 1, 2, "no", nil, nil)
	assert.NoError(t, err)
	assert.True(t, rsvps.cleared)
}

func TestEventService_UpdateRSVP_Concert_RequiresInstrument(t *testing.T) {
	rsvps := &stubRSVPRepoE{rsvp: &models.RSVPRow{ID: 10, State: "unanswered"}}
	svc := services.EventService{
		Events: &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "concert"}},
		RSVPs:  rsvps,
	}
	err := svc.UpdateRSVP(nil, 1, 2, "yes", nil, nil)
	assert.ErrorIs(t, err, services.ErrInstrumentRequired)
}

func TestEventService_UpdateRSVP_RSVPNotFound(t *testing.T) {
	rsvps := &stubRSVPRepoE{rsvp: nil}
	svc := services.EventService{
		Events: &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "rehearsal"}},
		RSVPs:  rsvps,
	}
	err := svc.UpdateRSVP(nil, 1, 99, "yes", nil, nil)
	assert.ErrorIs(t, err, services.ErrRSVPNotFound)
}

func TestEventService_UpdateRSVP_Other_SavesFieldResponses(t *testing.T) {
	rsvps := &stubRSVPRepoE{rsvp: &models.RSVPRow{ID: 10, State: "unanswered"}}
	svc := services.EventService{
		Events: &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "other"}},
		RSVPs:  rsvps,
	}
	fieldResponses := []services.FieldResponseInput{{FieldID: 1, Value: "3"}}
	err := svc.UpdateRSVP(nil, 1, 2, "yes", nil, fieldResponses)
	assert.NoError(t, err)
}

// --- AC-M5: Field edit and delete blocked when responses exist ---

func TestEventService_UpdateField_BlockedWhenResponsesExist(t *testing.T) {
	svc := services.EventService{
		Events: &stubEventRepoE{countResp: 1},
		RSVPs:  &stubRSVPRepoE{},
	}
	err := svc.UpdateField(nil, 5, "Label", "text", false, 1, nil)
	assert.ErrorIs(t, err, services.ErrFieldHasResponses)
}

func TestEventService_DeleteField_BlockedWhenResponsesExist(t *testing.T) {
	svc := services.EventService{
		Events: &stubEventRepoE{countResp: 1},
		RSVPs:  &stubRSVPRepoE{},
	}
	err := svc.DeleteField(nil, 5)
	assert.ErrorIs(t, err, services.ErrFieldHasResponses)
}

func TestEventService_UpdateField_AllowedWhenNoResponses(t *testing.T) {
	svc := services.EventService{
		Events: &stubEventRepoE{countResp: 0},
		RSVPs:  &stubRSVPRepoE{},
	}
	err := svc.UpdateField(nil, 5, "Label", "text", false, 1, nil)
	assert.NoError(t, err)
}

func TestEventService_DeleteField_AllowedWhenNoResponses(t *testing.T) {
	svc := services.EventService{
		Events: &stubEventRepoE{countResp: 0},
		RSVPs:  &stubRSVPRepoE{},
	}
	err := svc.DeleteField(nil, 5)
	assert.NoError(t, err)
}

// --- AddField: only for other events ---

func TestEventService_AddField_RejectsNonOtherEvent(t *testing.T) {
	svc := services.EventService{
		Events: &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "concert"}},
		RSVPs:  &stubRSVPRepoE{},
	}
	err := svc.AddField(nil, 1, "Label", "text", false, 1, nil)
	assert.ErrorIs(t, err, services.ErrFieldOnlyForOther)
}

func TestEventService_AddField_AcceptsOtherEvent(t *testing.T) {
	svc := services.EventService{
		Events: &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "other"}, fieldID: 10},
		RSVPs:  &stubRSVPRepoE{},
	}
	err := svc.AddField(nil, 1, "Label", "text", false, 1, nil)
	assert.NoError(t, err)
}

// --- SeedRSVPsForAccount delegates to repo ---

func TestEventService_SeedRSVPsForAccount_CallsRepo(t *testing.T) {
	rsvps := &stubRSVPRepoE{}
	svc := services.EventService{
		Events: &stubEventRepoE{},
		RSVPs:  rsvps,
	}
	err := svc.SeedRSVPsForAccount(nil, 42)
	assert.NoError(t, err)
	assert.True(t, rsvps.seeded)
}
