package services

import (
	"errors"
	"fmt"
	"time"

	"ohmontrouge/webapp/models"

	"github.com/gobuffalo/pop/v6"
)

var (
	ErrEventNotFound      = errors.New("événement introuvable")
	ErrRSVPNotFound       = errors.New("participation introuvable")
	ErrFieldHasResponses  = errors.New("ce champ a déjà des réponses et ne peut pas être modifié ou supprimé")
	ErrFieldOnlyForOther  = errors.New("les champs personnalisés ne sont disponibles que pour les événements de type « autre »")
	ErrInstrumentRequired = errors.New("l'instrument est requis pour un RSVP « présent » à un concert")
	ErrEventFieldNotFound = errors.New("champ introuvable")
)

// FieldChoiceInput is the input type for creating or replacing field choices.
type FieldChoiceInput struct {
	Label    string
	Position int
}

// FieldResponseInput is one custom field answer submitted with an RSVP.
type FieldResponseInput struct {
	FieldID int64
	Value   string
}

// EventSummaryDTO is the list-view representation of an event.
type EventSummaryDTO struct {
	ID          int64
	Name        string
	EventType   string // "concert" | "rehearsal" | "other"
	Datetime    time.Time
	RSVPState   string // viewer's own state; "" if no RSVP record
	Description string
}

// OwnRSVPDTO carries the viewer's own RSVP state for an event detail page.
type OwnRSVPDTO struct {
	ID             int64
	State          string
	InstrumentID   *int64
	FieldResponses []FieldResponseDTO
}

// FieldValue returns the raw stored value for fieldID, or "" if not found.
func (o *OwnRSVPDTO) FieldValue(fieldID int64) string {
	for _, fr := range o.FieldResponses {
		if fr.FieldID == fieldID {
			return fr.Value
		}
	}
	return ""
}

// RSVPRowDTO is one row in the full RSVP list on an event detail page.
type RSVPRowDTO struct {
	AccountID      int64
	DisplayName    string // "Nom Prénom" or anonymized placeholder
	State          string
	InstrumentID   int64  // 0 if none
	InstrumentName string // concert instrument for yes RSVPs
	MainInstrument string // musician's primary instrument (for pupitre)
	FieldResponses []FieldResponseDTO
}

// FieldValueMap returns a map of fieldID-as-string → raw value for use in templates.
func (r RSVPRowDTO) FieldValueMap() map[string]string {
	m := make(map[string]string, len(r.FieldResponses))
	for _, fr := range r.FieldResponses {
		m[fmt.Sprintf("%d", fr.FieldID)] = fr.Value
	}
	return m
}

// PupitreRowDTO is one row in the by-instrument headcount table (concerts).
type PupitreRowDTO struct {
	InstrumentName string
	Yes            int
	Maybe          int
	No             int
	Unanswered     int
}

// FieldResponseDTO is a musician's answer to one custom event field.
type FieldResponseDTO struct {
	FieldID      int64
	FieldLabel   string
	FieldType    string
	Value        string // raw stored value
	DisplayValue string // human-readable (choice label for choice-type fields)
}

// EventFieldDTO is a custom field definition on an event.
type EventFieldDTO struct {
	ID        int64
	EventID   int64
	Label     string
	FieldType string // "choice" | "integer" | "text"
	Required  bool
	Position  int
	Choices   []EventFieldChoiceDTO
}

// EventFieldChoiceDTO is one selectable option for a choice-type field.
type EventFieldChoiceDTO struct {
	ID       int64
	Label    string
	Position int
}

// EventDetailDTO is the full representation of an event for the detail page.
type EventDetailDTO struct {
	ID          int64
	Name        string
	EventType   string
	Datetime    time.Time
	Description string
	OwnRSVP     *OwnRSVPDTO
	RSVPs       []RSVPRowDTO
	Pupitre     []PupitreRowDTO // computed from RSVPs; only meaningful for concerts
	Fields      []EventFieldDTO // custom fields; only for other-type events
}

// EventManager is the interface EventsHandler depends on for all event operations.
type EventManager interface {
	ListForMember(tx *pop.Connection, accountID int64) ([]EventSummaryDTO, error)
	ListAll(tx *pop.Connection, accountID int64) ([]EventSummaryDTO, error)
	GetDetail(tx *pop.Connection, eventID, accountID int64) (*EventDetailDTO, error)
	Create(tx *pop.Connection, name, eventType, description string, datetime time.Time) error
	Update(tx *pop.Connection, id int64, name, eventType, description string, datetime time.Time) error
	Delete(tx *pop.Connection, id int64) error
	UpdateRSVP(tx *pop.Connection, eventID, accountID int64, state string, instrumentID *int64, fieldResponses []FieldResponseInput) error
	GetField(tx *pop.Connection, fieldID int64) (*EventFieldDTO, error)
	AddField(tx *pop.Connection, eventID int64, label, fieldType string, required bool, position int, choices []FieldChoiceInput) error
	UpdateField(tx *pop.Connection, fieldID int64, label, fieldType string, required bool, position int, choices []FieldChoiceInput) error
	DeleteField(tx *pop.Connection, fieldID int64) error
	SeedRSVPsForAccount(tx *pop.Connection, accountID int64) error
}

// EventService implements domain logic for events, RSVPs, and custom fields.
// It is the sole owner of the Event Coordination bounded context.
type EventService struct {
	Events EventRepository
	RSVPs  RSVPRepository
}

// ListForMember returns all upcoming events (today included) with the viewer's own RSVP state.
// Used for the /tableau-de-bord dashboard.
func (s EventService) ListForMember(tx *pop.Connection, accountID int64) ([]EventSummaryDTO, error) {
	rows, err := s.Events.ListUpcoming(tx, accountID)
	if err != nil {
		return nil, err
	}
	return toSummaryDTOs(rows), nil
}

// ListAll returns all events (past and future) with the viewer's own RSVP state.
// Used for the /evenements full list.
func (s EventService) ListAll(tx *pop.Connection, accountID int64) ([]EventSummaryDTO, error) {
	rows, err := s.Events.ListAll(tx, accountID)
	if err != nil {
		return nil, err
	}
	return toSummaryDTOs(rows), nil
}

func toSummaryDTOs(rows []models.EventListRow) []EventSummaryDTO {
	dtos := make([]EventSummaryDTO, len(rows))
	for i, r := range rows {
		state := ""
		if r.RSVPState.Valid {
			state = r.RSVPState.String
		}
		desc := ""
		if r.Description.Valid {
			desc = r.Description.String
		}
		dtos[i] = EventSummaryDTO{
			ID:          r.ID,
			Name:        r.Name,
			EventType:   r.EventType,
			Datetime:    r.Datetime,
			RSVPState:   state,
			Description: desc,
		}
	}
	return dtos
}

// GetDetail returns the full event detail including RSVPs and custom fields.
func (s EventService) GetDetail(tx *pop.Connection, eventID, accountID int64) (*EventDetailDTO, error) {
	event, err := s.Events.GetByID(tx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, ErrEventNotFound
	}

	rsvpRows, err := s.RSVPs.ListForEvent(tx, eventID)
	if err != nil {
		return nil, err
	}

	// Own RSVP
	ownRSVP, err := s.RSVPs.GetByAccountAndEvent(tx, accountID, eventID)
	if err != nil {
		return nil, err
	}

	// Fields (other events only)
	var fields []EventFieldDTO
	if event.EventType == "other" {
		fields, err = s.loadFields(tx, eventID)
		if err != nil {
			return nil, err
		}
	}

	// Field responses for yes RSVPs (other events only)
	var responsesByRSVP map[int64][]FieldResponseDTO
	if event.EventType == "other" {
		responsesByRSVP, err = s.loadFieldResponses(tx, eventID, fields)
		if err != nil {
			return nil, err
		}
	}

	// Build RSVPRowDTOs
	rsvps := make([]RSVPRowDTO, len(rsvpRows))
	for i, r := range rsvpRows {
		var instrID int64
		if r.InstrumentID.Valid {
			instrID = r.InstrumentID.Int64
		}
		rsvps[i] = RSVPRowDTO{
			AccountID:      r.AccountID,
			DisplayName:    displayName(r),
			State:          r.State,
			InstrumentID:   instrID,
			InstrumentName: r.RSVPInstrumentName.String,
			MainInstrument: r.MainInstrumentName,
			FieldResponses: responsesByRSVP[r.RSVPID],
		}
	}

	// Pupitre headcount (all event types — useful for setup planning)
	pupitre := buildPupitre(rsvpRows)

	desc := ""
	if event.Description.Valid {
		desc = event.Description.String
	}

	dto := &EventDetailDTO{
		ID:          event.ID,
		Name:        event.Name,
		EventType:   event.EventType,
		Datetime:    event.Datetime,
		Description: desc,
		RSVPs:       rsvps,
		Pupitre:     pupitre,
		Fields:      fields,
	}
	if ownRSVP != nil {
		var instrID *int64
		if ownRSVP.InstrumentID.Valid {
			v := ownRSVP.InstrumentID.Int64
			instrID = &v
		}
		dto.OwnRSVP = &OwnRSVPDTO{
			ID:             ownRSVP.ID,
			State:          ownRSVP.State,
			InstrumentID:   instrID,
			FieldResponses: responsesByRSVP[ownRSVP.ID],
		}
	}
	return dto, nil
}

func displayName(r models.RSVPListRow) string {
	if r.FirstName.Valid && r.LastName.Valid {
		return fmt.Sprintf("%s %s", r.LastName.String, r.FirstName.String)
	}
	if r.AnonymizationToken.Valid {
		return "Musicien " + r.AnonymizationToken.String[:8]
	}
	return "Compte inconnu"
}

func (s EventService) loadFields(tx *pop.Connection, eventID int64) ([]EventFieldDTO, error) {
	fieldRows, err := s.Events.ListFields(tx, eventID)
	if err != nil {
		return nil, err
	}
	fields := make([]EventFieldDTO, len(fieldRows))
	for i, f := range fieldRows {
		var choices []EventFieldChoiceDTO
		if f.FieldType == "choice" {
			choiceRows, err := s.Events.ListFieldChoices(tx, f.ID)
			if err != nil {
				return nil, err
			}
			choices = make([]EventFieldChoiceDTO, len(choiceRows))
			for j, c := range choiceRows {
				choices[j] = EventFieldChoiceDTO{ID: c.ID, Label: c.Label, Position: c.Position}
			}
		}
		fields[i] = EventFieldDTO{
			ID:        f.ID,
			EventID:   f.EventID,
			Label:     f.Label,
			FieldType: f.FieldType,
			Required:  f.Required,
			Position:  f.Position,
			Choices:   choices,
		}
	}
	return fields, nil
}

func (s EventService) loadFieldResponses(tx *pop.Connection, eventID int64, fields []EventFieldDTO) (map[int64][]FieldResponseDTO, error) {
	responseRows, err := s.RSVPs.ListFieldResponses(tx, eventID)
	if err != nil {
		return nil, err
	}

	// Build choice label lookup: choiceID (as string) → label
	choiceLabels := map[string]string{}
	for _, f := range fields {
		for _, c := range f.Choices {
			choiceLabels[fmt.Sprintf("%d", c.ID)] = c.Label
		}
	}

	result := map[int64][]FieldResponseDTO{}
	for _, r := range responseRows {
		display := r.Value
		if r.FieldType == "choice" {
			if label, ok := choiceLabels[r.Value]; ok {
				display = label
			}
		}
		result[r.RSVPID] = append(result[r.RSVPID], FieldResponseDTO{
			FieldID:      r.EventFieldID,
			FieldLabel:   r.FieldLabel,
			FieldType:    r.FieldType,
			Value:        r.Value,
			DisplayValue: display,
		})
	}
	return result, nil
}

// buildPupitre computes per-instrument headcounts. For yes RSVPs, uses the selected concert
// instrument; for all other states, uses the musician's main instrument.
func buildPupitre(rsvps []models.RSVPListRow) []PupitreRowDTO {
	counts := map[string]*PupitreRowDTO{}
	order := []string{}

	for _, r := range rsvps {
		inst := r.MainInstrumentName
		if r.State == "yes" && r.RSVPInstrumentName.Valid && r.RSVPInstrumentName.String != "" {
			inst = r.RSVPInstrumentName.String
		}
		if _, ok := counts[inst]; !ok {
			counts[inst] = &PupitreRowDTO{InstrumentName: inst}
			order = append(order, inst)
		}
		switch r.State {
		case "yes":
			counts[inst].Yes++
		case "maybe":
			counts[inst].Maybe++
		case "no":
			counts[inst].No++
		default:
			counts[inst].Unanswered++
		}
	}

	rows := make([]PupitreRowDTO, len(order))
	for i, inst := range order {
		rows[i] = *counts[inst]
	}
	return rows
}

// Create creates an event and seeds RSVPs for all active accounts.
func (s EventService) Create(tx *pop.Connection, name, eventType, description string, datetime time.Time) error {
	eventID, err := s.Events.Create(tx, name, eventType, description, datetime)
	if err != nil {
		return err
	}
	return s.RSVPs.SeedForEvent(tx, eventID)
}

// Update saves event changes and applies type-change effects atomically.
func (s EventService) Update(tx *pop.Connection, id int64, name, eventType, description string, datetime time.Time) error {
	event, err := s.Events.GetByID(tx, id)
	if err != nil {
		return err
	}
	if event == nil {
		return ErrEventNotFound
	}
	oldType := event.EventType

	if err := s.Events.Update(tx, id, name, eventType, description, datetime); err != nil {
		return err
	}

	if oldType == eventType {
		return nil
	}

	// Apply type-change effects per spec table
	switch {
	case eventType == "concert" && oldType == "rehearsal":
		return s.RSVPs.ResetYesRSVPs(tx, id)

	case eventType == "concert" && oldType == "other":
		if err := s.Events.DeleteFields(tx, id); err != nil {
			return err
		}
		return s.RSVPs.ResetYesRSVPs(tx, id)

	case eventType == "rehearsal" && oldType == "other":
		return s.Events.DeleteFields(tx, id)

	case oldType == "concert": // to rehearsal or other from concert
		return s.RSVPs.ClearInstruments(tx, id)
	}

	return nil
}

// Delete removes an event; DB FK cascade removes all RSVPs.
func (s EventService) Delete(tx *pop.Connection, id int64) error {
	return s.Events.Delete(tx, id)
}

// UpdateRSVP updates a musician's RSVP state with instrument validation, field response clearing, and field response saving.
func (s EventService) UpdateRSVP(tx *pop.Connection, eventID, accountID int64, state string, instrumentID *int64, fieldResponses []FieldResponseInput) error {
	event, err := s.Events.GetByID(tx, eventID)
	if err != nil {
		return err
	}
	if event == nil {
		return ErrEventNotFound
	}

	rsvp, err := s.RSVPs.GetByAccountAndEvent(tx, accountID, eventID)
	if err != nil {
		return err
	}
	if rsvp == nil {
		return ErrRSVPNotFound
	}

	// Concerts require an instrument when RSVPing yes
	if event.EventType == "concert" && state == "yes" && instrumentID == nil {
		return ErrInstrumentRequired
	}

	// Instrument is only stored for concert yes RSVPs; clear otherwise
	var instrPtr *int64
	if event.EventType == "concert" && state == "yes" {
		instrPtr = instrumentID
	}

	// Previous state was yes: clear so optional fields can be removed and re-add is clean
	if rsvp.State == "yes" {
		if err := s.RSVPs.ClearFieldResponses(tx, rsvp.ID); err != nil {
			return err
		}
	}

	if err := s.RSVPs.Update(tx, rsvp.ID, state, instrPtr); err != nil {
		return err
	}

	// Save field responses when yes (upsert — AddFieldResponse does INSERT ON CONFLICT UPDATE)
	if state == "yes" {
		for _, fr := range fieldResponses {
			if err := s.RSVPs.AddFieldResponse(tx, rsvp.ID, fr.FieldID, fr.Value); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetField returns the DTO for a single event field.
func (s EventService) GetField(tx *pop.Connection, fieldID int64) (*EventFieldDTO, error) {
	f, err := s.Events.GetFieldByID(tx, fieldID)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, ErrEventFieldNotFound
	}
	var choices []EventFieldChoiceDTO
	if f.FieldType == "choice" {
		choiceRows, err := s.Events.ListFieldChoices(tx, f.ID)
		if err != nil {
			return nil, err
		}
		choices = make([]EventFieldChoiceDTO, len(choiceRows))
		for i, c := range choiceRows {
			choices[i] = EventFieldChoiceDTO{ID: c.ID, Label: c.Label, Position: c.Position}
		}
	}
	return &EventFieldDTO{
		ID:        f.ID,
		EventID:   f.EventID,
		Label:     f.Label,
		FieldType: f.FieldType,
		Required:  f.Required,
		Position:  f.Position,
		Choices:   choices,
	}, nil
}

// AddField adds a custom field to an other-type event.
func (s EventService) AddField(tx *pop.Connection, eventID int64, label, fieldType string, required bool, position int, choices []FieldChoiceInput) error {
	event, err := s.Events.GetByID(tx, eventID)
	if err != nil {
		return err
	}
	if event == nil {
		return ErrEventNotFound
	}
	if event.EventType != "other" {
		return ErrFieldOnlyForOther
	}

	fieldID, err := s.Events.AddField(tx, eventID, label, fieldType, required, position)
	if err != nil {
		return err
	}
	return s.addChoices(tx, fieldID, fieldType, choices)
}

// UpdateField updates a field's properties. Blocked when any responses exist.
func (s EventService) UpdateField(tx *pop.Connection, fieldID int64, label, fieldType string, required bool, position int, choices []FieldChoiceInput) error {
	count, err := s.Events.CountFieldResponses(tx, fieldID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrFieldHasResponses
	}

	if err := s.Events.UpdateField(tx, fieldID, label, fieldType, required, position); err != nil {
		return err
	}

	// Replace choices when field type is choice
	if err := s.Events.DeleteFieldChoices(tx, fieldID); err != nil {
		return err
	}
	return s.addChoices(tx, fieldID, fieldType, choices)
}

// addChoices inserts choice rows for a field when fieldType is "choice".
func (s EventService) addChoices(tx *pop.Connection, fieldID int64, fieldType string, choices []FieldChoiceInput) error {
	if fieldType == "choice" {
		for _, c := range choices {
			if _, err := s.Events.AddFieldChoice(tx, fieldID, c.Label, c.Position); err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteField removes a field and its choices. Blocked when any responses exist.
func (s EventService) DeleteField(tx *pop.Connection, fieldID int64) error {
	count, err := s.Events.CountFieldResponses(tx, fieldID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrFieldHasResponses
	}
	return s.Events.DeleteField(tx, fieldID)
}

// SeedRSVPsForAccount creates unanswered RSVP records for all future events.
// Called by AccountService.CompleteInvite when a musician activates their account.
func (s EventService) SeedRSVPsForAccount(tx *pop.Connection, accountID int64) error {
	return s.RSVPs.SeedForAccount(tx, accountID)
}
