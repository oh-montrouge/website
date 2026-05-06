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
	event           *models.EventDetailRow
	eventID         int64
	err             error
	fieldID         int64
	fieldErr        error
	field           *models.EventFieldRow
	countResp       int
	deleted         bool
	updated         bool
	listRows        []models.EventListRow
	fieldRows       []models.EventFieldRow
	fieldChoiceRows []models.EventFieldChoiceRow
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
	return s.listRows, s.err
}
func (s *stubEventRepoE) ListAll(_ *pop.Connection, _ int64) ([]models.EventListRow, error) {
	return s.listRows, s.err
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
	return s.fieldRows, s.err
}
func (s *stubEventRepoE) ListFieldChoices(_ *pop.Connection, _ int64) ([]models.EventFieldChoiceRow, error) {
	return s.fieldChoiceRows, s.err
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
	rsvp              *models.RSVPRow
	err               error
	seeded            bool
	cleared           bool
	resetYes          bool
	clearInstrument   bool
	rsvpRows          []models.RSVPListRow
	fieldResponseRows []models.RSVPFieldResponseRow
	listErr           error
	ownRSVPErr        error
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
	return s.rsvp, s.ownRSVPErr
}
func (s *stubRSVPRepoE) Update(_ *pop.Connection, _ int64, _ string, _ *int64) error { return s.err }
func (s *stubRSVPRepoE) DeleteByAccount(_ *pop.Connection, _ int64) error            { return s.err }
func (s *stubRSVPRepoE) ClearFieldResponses(_ *pop.Connection, _ int64) error {
	s.cleared = true
	return s.err
}
func (s *stubRSVPRepoE) ListForEvent(_ *pop.Connection, _ int64) ([]models.RSVPListRow, error) {
	return s.rsvpRows, s.listErr
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
	return s.fieldResponseRows, s.err
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

// --- OwnRSVPDTO.FieldValue ---

func TestOwnRSVPDTO_FieldValue_Found(t *testing.T) {
	dto := &services.OwnRSVPDTO{
		FieldResponses: []services.FieldResponseDTO{{FieldID: 5, Value: "abc"}},
	}
	assert.Equal(t, "abc", dto.FieldValue(5))
}

func TestOwnRSVPDTO_FieldValue_NotFound(t *testing.T) {
	dto := &services.OwnRSVPDTO{
		FieldResponses: []services.FieldResponseDTO{{FieldID: 5, Value: "abc"}},
	}
	assert.Equal(t, "", dto.FieldValue(99))
}

// --- RSVPRowDTO.FieldValueMap ---

func TestRSVPRowDTO_FieldValueMap_Empty(t *testing.T) {
	row := services.RSVPRowDTO{}
	assert.Empty(t, row.FieldValueMap())
}

func TestRSVPRowDTO_FieldValueMap_Populated(t *testing.T) {
	row := services.RSVPRowDTO{
		FieldResponses: []services.FieldResponseDTO{
			{FieldID: 10, Value: "hello"},
			{FieldID: 20, Value: "world"},
		},
	}
	m := row.FieldValueMap()
	assert.Equal(t, "hello", m["10"])
	assert.Equal(t, "world", m["20"])
}

// --- ListForMember ---

func TestEventService_ListForMember_Success(t *testing.T) {
	rows := []models.EventListRow{
		{ID: 1, Name: "Concert", EventType: "concert", RSVPState: nulls.NewString("yes")},
	}
	svc := services.EventService{
		Events: &stubEventRepoE{listRows: rows},
		RSVPs:  &stubRSVPRepoE{},
	}
	dtos, err := svc.ListForMember(nil, 42)
	assert.NoError(t, err)
	assert.Len(t, dtos, 1)
	assert.Equal(t, int64(1), dtos[0].ID)
	assert.Equal(t, "yes", dtos[0].RSVPState)
}

func TestEventService_ListForMember_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := services.EventService{
		Events: &stubEventRepoE{err: repoErr},
		RSVPs:  &stubRSVPRepoE{},
	}
	_, err := svc.ListForMember(nil, 42)
	assert.ErrorIs(t, err, repoErr)
}

// --- ListAll ---

func TestEventService_ListAll_Success(t *testing.T) {
	rows := []models.EventListRow{
		{ID: 2, Name: "Répétition", EventType: "rehearsal"},
	}
	svc := services.EventService{
		Events: &stubEventRepoE{listRows: rows},
		RSVPs:  &stubRSVPRepoE{},
	}
	dtos, err := svc.ListAll(nil, 42)
	assert.NoError(t, err)
	assert.Len(t, dtos, 1)
	assert.Equal(t, int64(2), dtos[0].ID)
}

func TestEventService_ListAll_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := services.EventService{
		Events: &stubEventRepoE{err: repoErr},
		RSVPs:  &stubRSVPRepoE{},
	}
	_, err := svc.ListAll(nil, 42)
	assert.ErrorIs(t, err, repoErr)
}

// --- GetDetail ---

func TestEventService_GetDetail_EventNotFound(t *testing.T) {
	svc := services.EventService{
		Events: &stubEventRepoE{event: nil},
		RSVPs:  &stubRSVPRepoE{},
	}
	_, err := svc.GetDetail(nil, 1, 2)
	assert.ErrorIs(t, err, services.ErrEventNotFound)
}

func TestEventService_GetDetail_GetByIDError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := services.EventService{
		Events: &stubEventRepoE{err: repoErr},
		RSVPs:  &stubRSVPRepoE{},
	}
	_, err := svc.GetDetail(nil, 1, 2)
	assert.ErrorIs(t, err, repoErr)
}

func TestEventService_GetDetail_RSVPsError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := services.EventService{
		Events: &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "rehearsal"}},
		RSVPs:  &stubRSVPRepoE{listErr: repoErr},
	}
	_, err := svc.GetDetail(nil, 1, 2)
	assert.ErrorIs(t, err, repoErr)
}

func TestEventService_GetDetail_OwnRSVPError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := services.EventService{
		Events: &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "rehearsal"}},
		RSVPs:  &stubRSVPRepoE{ownRSVPErr: repoErr},
	}
	_, err := svc.GetDetail(nil, 1, 2)
	assert.ErrorIs(t, err, repoErr)
}

func TestEventService_GetDetail_ConcertEvent_NoFields(t *testing.T) {
	svc := services.EventService{
		Events: &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "concert"}},
		RSVPs:  &stubRSVPRepoE{},
	}
	dto, err := svc.GetDetail(nil, 1, 2)
	assert.NoError(t, err)
	assert.Empty(t, dto.Fields)
}

func TestEventService_GetDetail_OtherEvent_LoadsFields(t *testing.T) {
	svc := services.EventService{
		Events: &stubEventRepoE{
			event:     &models.EventDetailRow{ID: 1, EventType: "other"},
			fieldRows: []models.EventFieldRow{{ID: 10, EventID: 1, Label: "Note", FieldType: "text"}},
		},
		RSVPs: &stubRSVPRepoE{},
	}
	dto, err := svc.GetDetail(nil, 1, 2)
	assert.NoError(t, err)
	assert.Len(t, dto.Fields, 1)
	assert.Equal(t, "Note", dto.Fields[0].Label)
}

func TestEventService_GetDetail_OtherEvent_LoadsFieldResponses(t *testing.T) {
	svc := services.EventService{
		Events: &stubEventRepoE{
			event:     &models.EventDetailRow{ID: 1, EventType: "other"},
			fieldRows: []models.EventFieldRow{{ID: 10, EventID: 1, Label: "Note", FieldType: "text"}},
		},
		RSVPs: &stubRSVPRepoE{
			rsvpRows: []models.RSVPListRow{
				{RSVPID: 100, AccountID: 5, State: "yes", MainInstrumentName: "Violon",
					FirstName: nulls.NewString("Alice"), LastName: nulls.NewString("Dupont")},
			},
			fieldResponseRows: []models.RSVPFieldResponseRow{
				{RSVPID: 100, EventFieldID: 10, FieldLabel: "Note", FieldType: "text", Value: "3"},
			},
		},
	}
	dto, err := svc.GetDetail(nil, 1, 2)
	assert.NoError(t, err)
	assert.Len(t, dto.RSVPs, 1)
	assert.Len(t, dto.RSVPs[0].FieldResponses, 1)
	assert.Equal(t, "3", dto.RSVPs[0].FieldResponses[0].Value)
}

func TestEventService_GetDetail_WithOwnRSVP(t *testing.T) {
	instrID := int64(5)
	svc := services.EventService{
		Events: &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "rehearsal"}},
		RSVPs: &stubRSVPRepoE{
			rsvp: &models.RSVPRow{ID: 100, State: "yes", InstrumentID: nulls.NewInt64(instrID)},
		},
	}
	dto, err := svc.GetDetail(nil, 1, 2)
	assert.NoError(t, err)
	assert.NotNil(t, dto.OwnRSVP)
	assert.Equal(t, "yes", dto.OwnRSVP.State)
	assert.Equal(t, instrID, *dto.OwnRSVP.InstrumentID)
}

func TestEventService_GetDetail_WithoutOwnRSVP(t *testing.T) {
	svc := services.EventService{
		Events: &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "rehearsal"}},
		RSVPs:  &stubRSVPRepoE{rsvp: nil},
	}
	dto, err := svc.GetDetail(nil, 1, 2)
	assert.NoError(t, err)
	assert.Nil(t, dto.OwnRSVP)
}

func TestEventService_GetDetail_NullableDescription(t *testing.T) {
	svc := services.EventService{
		Events: &stubEventRepoE{event: &models.EventDetailRow{ID: 1, EventType: "rehearsal"}},
		RSVPs:  &stubRSVPRepoE{},
	}
	dto, err := svc.GetDetail(nil, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, "", dto.Description)
}

// --- GetField ---

func TestEventService_GetField_NotFound(t *testing.T) {
	svc := services.EventService{
		Events: &stubEventRepoE{field: nil},
		RSVPs:  &stubRSVPRepoE{},
	}
	_, err := svc.GetField(nil, 5)
	assert.ErrorIs(t, err, services.ErrEventFieldNotFound)
}

func TestEventService_GetField_NonChoiceType(t *testing.T) {
	svc := services.EventService{
		Events: &stubEventRepoE{field: &models.EventFieldRow{ID: 5, FieldType: "text"}},
		RSVPs:  &stubRSVPRepoE{},
	}
	dto, err := svc.GetField(nil, 5)
	assert.NoError(t, err)
	assert.Empty(t, dto.Choices)
}

func TestEventService_GetField_ChoiceType(t *testing.T) {
	svc := services.EventService{
		Events: &stubEventRepoE{
			field:           &models.EventFieldRow{ID: 5, FieldType: "choice"},
			fieldChoiceRows: []models.EventFieldChoiceRow{{ID: 1, Label: "Oui", Position: 1}},
		},
		RSVPs: &stubRSVPRepoE{},
	}
	dto, err := svc.GetField(nil, 5)
	assert.NoError(t, err)
	assert.Len(t, dto.Choices, 1)
	assert.Equal(t, "Oui", dto.Choices[0].Label)
}

func TestEventService_GetField_ChoiceTypeError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := services.EventService{
		Events: &stubEventRepoE{
			field: &models.EventFieldRow{ID: 5, FieldType: "choice"},
			err:   repoErr,
		},
		RSVPs: &stubRSVPRepoE{},
	}
	_, err := svc.GetField(nil, 5)
	assert.ErrorIs(t, err, repoErr)
}
