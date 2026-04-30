package actions

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ohmontrouge/webapp/services"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
)

// EventsHandler handles event listing, detail, RSVP, and admin event management.
type EventsHandler struct {
	Events      services.EventManager
	Instruments services.InstrumentRepository
	Membership  services.MusicianProfileManager
}

// Dashboard renders /tableau-de-bord — upcoming events with own RSVP state.
func (h EventsHandler) Dashboard(c buffalo.Context) error {
	account := c.Value("current_account").(*services.AccountDTO)
	tx := c.Value("tx").(*pop.Connection)
	events, err := h.Events.ListForMember(tx, account.ID)
	if err != nil {
		return err
	}
	c.Set("events", events)
	c.Set("eventsEmpty", len(events) == 0)
	return c.Render(http.StatusOK, r.HTML("events/index.plush.html"))
}

// Index renders /evenements — all events with own RSVP state.
func (h EventsHandler) Index(c buffalo.Context) error {
	account := c.Value("current_account").(*services.AccountDTO)
	tx := c.Value("tx").(*pop.Connection)
	events, err := h.Events.ListAll(tx, account.ID)
	if err != nil {
		return err
	}
	c.Set("events", events)
	c.Set("eventsEmpty", len(events) == 0)
	c.Set("dashboardView", false)
	return c.Render(http.StatusOK, r.HTML("events/index.plush.html"))
}

// loadEventDetail fetches an event detail by id, returning 404 on not found.
func (h EventsHandler) loadEventDetail(c buffalo.Context, tx *pop.Connection, id, accountID int64) (*services.EventDetailDTO, error) {
	detail, err := h.Events.GetDetail(tx, id, accountID)
	if err != nil {
		if errors.Is(err, services.ErrEventNotFound) {
			return nil, c.Error(http.StatusNotFound, err)
		}
		return nil, err
	}
	return detail, nil
}

// Show renders /evenements/{id} — event detail with full RSVP list.
func (h EventsHandler) Show(c buffalo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	account := c.Value("current_account").(*services.AccountDTO)
	tx := c.Value("tx").(*pop.Connection)

	detail, err := h.loadEventDetail(c, tx, id, account.ID)
	if err != nil {
		return err
	}

	// Instrument list for the concert RSVP dropdown
	instruments, err := h.Instruments.List(tx)
	if err != nil {
		return err
	}

	// Pre-selected instrument: own RSVP instrument if set, otherwise main instrument.
	selectedInstrumentID := int64(0)
	if detail.EventType == "concert" {
		if detail.OwnRSVP != nil && detail.OwnRSVP.InstrumentID != nil {
			selectedInstrumentID = *detail.OwnRSVP.InstrumentID
		} else {
			profile, err := h.Membership.GetProfile(tx, account.ID)
			if err == nil && profile != nil {
				selectedInstrumentID = profile.MainInstrumentID
			}
		}
	}

	c.Set("event", detail)
	c.Set("instruments", instruments)
	c.Set("selectedInstrumentID", selectedInstrumentID)
	return c.Render(http.StatusOK, r.HTML("events/show.plush.html"))
}

// UpdateRSVP handles POST /evenements/{id}/rsvp — update own RSVP state.
func (h EventsHandler) UpdateRSVP(c buffalo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusBadRequest, err)
	}
	account := c.Value("current_account").(*services.AccountDTO)
	tx := c.Value("tx").(*pop.Connection)

	state := strings.TrimSpace(c.Request().FormValue("state"))
	if state == "" {
		c.Flash().Add("danger", "La réponse est requise.")
		return c.Redirect(http.StatusSeeOther, "/evenements/%d", id)
	}

	var instrumentID *int64
	instrStr := strings.TrimSpace(c.Request().FormValue("instrument_id"))
	if instrStr != "" {
		v, err := strconv.ParseInt(instrStr, 10, 64)
		if err == nil && v > 0 {
			instrumentID = &v
		}
	}

	// Parse custom field responses: field_<ID> form values
	var fieldResponses []services.FieldResponseInput
	if err := c.Request().ParseForm(); err == nil {
		for key, vals := range c.Request().Form {
			if idStr, ok := strings.CutPrefix(key, "field_"); ok {
				fieldID, err := strconv.ParseInt(idStr, 10, 64)
				if err == nil && len(vals) > 0 {
					val := strings.TrimSpace(vals[0])
					if val != "" {
						fieldResponses = append(fieldResponses, services.FieldResponseInput{FieldID: fieldID, Value: val})
					}
				}
			}
		}
	}

	if err := h.Events.UpdateRSVP(tx, id, account.ID, state, instrumentID, fieldResponses); err != nil {
		if errors.Is(err, services.ErrInstrumentRequired) {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(http.StatusSeeOther, "/evenements/%d", id)
		}
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/evenements/%d", id)
}

// New renders /admin/evenements/nouveau — new event form.
func (h EventsHandler) New(c buffalo.Context) error {
	c.Set("formError", "")
	return c.Render(http.StatusOK, r.HTML("admin/events/new.plush.html"))
}

// Create handles POST /admin/evenements — create event and seed RSVPs.
func (h EventsHandler) Create(c buffalo.Context) error {
	name, dateStr, timeStr, eventType, formErr := parseEventForm(c)
	if formErr != "" {
		c.Set("formError", formErr)
		return c.Render(http.StatusUnprocessableEntity, r.HTML("admin/events/new.plush.html"))
	}

	dt, err := parseDatetime(dateStr, timeStr)
	if err != nil {
		c.Set("formError", "Date ou heure invalide.")
		return c.Render(http.StatusUnprocessableEntity, r.HTML("admin/events/new.plush.html"))
	}

	tx := c.Value("tx").(*pop.Connection)
	if err := h.Events.Create(tx, name, eventType, dt); err != nil {
		return err
	}

	c.Flash().Add("success", "Événement créé.")
	return c.Redirect(http.StatusSeeOther, "/admin/evenements")
}

// Edit renders /admin/evenements/{id}/modifier — edit event form.
func (h EventsHandler) Edit(c buffalo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	tx := c.Value("tx").(*pop.Connection)
	account := c.Value("current_account").(*services.AccountDTO)

	detail, err := h.loadEventDetail(c, tx, id, account.ID)
	if err != nil {
		return err
	}

	c.Set("event", detail)
	c.Set("formError", "")
	return c.Render(http.StatusOK, r.HTML("admin/events/edit.plush.html"))
}

// Update handles PUT /admin/evenements/{id} — save event edits with type-change effects.
func (h EventsHandler) Update(c buffalo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	name, dateStr, timeStr, eventType, formErr := parseEventForm(c)
	if formErr != "" {
		tx := c.Value("tx").(*pop.Connection)
		account := c.Value("current_account").(*services.AccountDTO)
		detail, _ := h.Events.GetDetail(tx, id, account.ID)
		c.Set("event", detail)
		c.Set("formError", formErr)
		return c.Render(http.StatusUnprocessableEntity, r.HTML("admin/events/edit.plush.html"))
	}

	dt, err := parseDatetime(dateStr, timeStr)
	if err != nil {
		tx := c.Value("tx").(*pop.Connection)
		account := c.Value("current_account").(*services.AccountDTO)
		detail, _ := h.Events.GetDetail(tx, id, account.ID)
		c.Set("event", detail)
		c.Set("formError", "Date ou heure invalide.")
		return c.Render(http.StatusUnprocessableEntity, r.HTML("admin/events/edit.plush.html"))
	}

	tx := c.Value("tx").(*pop.Connection)
	if err := h.Events.Update(tx, id, name, eventType, dt); err != nil {
		if errors.Is(err, services.ErrEventNotFound) {
			return c.Error(http.StatusNotFound, err)
		}
		return err
	}

	c.Flash().Add("success", "Événement mis à jour.")
	return c.Redirect(http.StatusSeeOther, "/admin/evenements")
}

// Delete handles DELETE /admin/evenements/{id} — delete event and all RSVPs.
func (h EventsHandler) Delete(c buffalo.Context) error {
	if c.Request().FormValue("confirmed") != "true" {
		c.Flash().Add("danger", "Suppression non confirmée.")
		return c.Redirect(http.StatusSeeOther, "/admin/evenements")
	}

	id, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	tx := c.Value("tx").(*pop.Connection)
	if err := h.Events.Delete(tx, id); err != nil {
		return err
	}

	c.Flash().Add("success", "Événement supprimé.")
	return c.Redirect(http.StatusSeeOther, "/admin/evenements")
}

// AddField handles POST /admin/evenements/{id}/champs.
func (h EventsHandler) AddField(c buffalo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusBadRequest, err)
	}

	label, fieldType, required, position, choices := parseFieldFormValues(c)

	tx := c.Value("tx").(*pop.Connection)
	if err := h.Events.AddField(tx, id, label, fieldType, required, position, choices); err != nil {
		if errors.Is(err, services.ErrFieldOnlyForOther) {
			c.Flash().Add("danger", err.Error())
		} else {
			return err
		}
	}

	return c.Redirect(http.StatusSeeOther, "/admin/evenements/%d/modifier", id)
}

// EditFieldForm renders /admin/evenements/{event_id}/champs/{field_id}/modifier.
func (h EventsHandler) EditFieldForm(c buffalo.Context) error {
	fieldID, err := strconv.ParseInt(c.Param("field_id"), 10, 64)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	eventID, err := parseID(c)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	tx := c.Value("tx").(*pop.Connection)
	field, err := h.Events.GetField(tx, fieldID)
	if err != nil {
		if errors.Is(err, services.ErrEventFieldNotFound) {
			return c.Error(http.StatusNotFound, err)
		}
		return err
	}

	c.Set("field", field)
	c.Set("eventID", eventID)
	c.Set("formError", "")
	return c.Render(http.StatusOK, r.HTML("admin/events/edit_field.plush.html"))
}

// UpdateField handles PUT /admin/evenements/{event_id}/champs/{field_id}.
func (h EventsHandler) UpdateField(c buffalo.Context) error {
	fieldID, eventID, err := parseFieldAndEventIDs(c)
	if err != nil {
		return err
	}

	label, fieldType, required, position, choices := parseFieldFormValues(c)

	tx := c.Value("tx").(*pop.Connection)
	if err := h.Events.UpdateField(tx, fieldID, label, fieldType, required, position, choices); err != nil {
		if errors.Is(err, services.ErrFieldHasResponses) {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(http.StatusSeeOther, "/admin/evenements/%d/modifier", eventID)
		}
		return err
	}

	c.Flash().Add("success", "Champ mis à jour.")
	return c.Redirect(http.StatusSeeOther, "/admin/evenements/%d/modifier", eventID)
}

// DeleteField handles DELETE /admin/evenements/{event_id}/champs/{field_id}.
func (h EventsHandler) DeleteField(c buffalo.Context) error {
	fieldID, eventID, err := parseFieldAndEventIDs(c)
	if err != nil {
		return err
	}

	tx := c.Value("tx").(*pop.Connection)
	if err := h.Events.DeleteField(tx, fieldID); err != nil {
		if errors.Is(err, services.ErrFieldHasResponses) {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(http.StatusSeeOther, "/admin/evenements/%d/modifier", eventID)
		}
		return err
	}

	c.Flash().Add("success", "Champ supprimé.")
	return c.Redirect(http.StatusSeeOther, "/admin/evenements/%d/modifier", eventID)
}

// AdminUpdateRSVP handles PATCH /admin/evenements/{id}/rsvp/{musician_id} — JSON endpoint.
// Returns {"ok":true} on success. Used by Alpine.js inline RSVP editing in the admin view.
func (h EventsHandler) AdminUpdateRSVP(c buffalo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid event id")
	}
	musicianID, err := strconv.ParseInt(c.Param("musician_id"), 10, 64)
	if err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid musician id")
	}

	var body struct {
		State          string `json:"state"`
		InstrumentID   *int64 `json:"instrument_id"`
		FieldResponses []struct {
			FieldID int64  `json:"field_id"`
			Value   string `json:"value"`
		} `json:"field_responses"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid request body")
	}

	fieldResponses := make([]services.FieldResponseInput, 0, len(body.FieldResponses))
	for _, fr := range body.FieldResponses {
		if fr.Value != "" {
			fieldResponses = append(fieldResponses, services.FieldResponseInput{FieldID: fr.FieldID, Value: fr.Value})
		}
	}

	tx := c.Value("tx").(*pop.Connection)
	if err := h.Events.UpdateRSVP(tx, id, musicianID, body.State, body.InstrumentID, fieldResponses); err != nil {
		if errors.Is(err, services.ErrInstrumentRequired) {
			return jsonError(c, http.StatusUnprocessableEntity, err.Error())
		}
		if errors.Is(err, services.ErrRSVPNotFound) {
			return jsonError(c, http.StatusNotFound, err.Error())
		}
		return err
	}

	return c.Render(http.StatusOK, r.JSON(map[string]bool{"ok": true}))
}

// --- helpers ---

// parseFieldAndEventIDs extracts field_id and the route :id from the request.
func parseFieldAndEventIDs(c buffalo.Context) (fieldID, eventID int64, err error) {
	fieldID, err = strconv.ParseInt(c.Param("field_id"), 10, 64)
	if err != nil {
		return 0, 0, c.Error(http.StatusBadRequest, err)
	}
	eventID, err = parseID(c)
	if err != nil {
		return 0, 0, c.Error(http.StatusBadRequest, err)
	}
	return fieldID, eventID, nil
}

// parseFieldFormValues extracts the shared field form fields from the request.
func parseFieldFormValues(c buffalo.Context) (label, fieldType string, required bool, position int, choices []services.FieldChoiceInput) {
	label = strings.TrimSpace(c.Request().FormValue("label"))
	fieldType = strings.TrimSpace(c.Request().FormValue("field_type"))
	required = c.Request().FormValue("required") == "true"
	positionStr := c.Request().FormValue("position")
	position, _ = strconv.Atoi(positionStr)
	choices = parseChoices(c)
	return
}

func parseEventForm(c buffalo.Context) (name, dateStr, timeStr, eventType, formErr string) {
	name = strings.TrimSpace(c.Request().FormValue("name"))
	dateStr = c.Request().FormValue("date")
	timeStr = c.Request().FormValue("time")
	eventType = strings.TrimSpace(c.Request().FormValue("event_type"))

	if name == "" {
		formErr = "Le nom est requis."
		return
	}
	validTypes := map[string]bool{"concert": true, "rehearsal": true, "other": true}
	if !validTypes[eventType] {
		formErr = "Le type d'événement est invalide."
		return
	}
	return
}

func parseDatetime(dateStr, timeStr string) (time.Time, error) {
	if timeStr == "" {
		timeStr = "00:00"
	}
	return time.Parse("2006-01-02 15:04", dateStr+" "+timeStr)
}

func parseChoices(c buffalo.Context) []services.FieldChoiceInput {
	labels := c.Request().Form["choice_label[]"]
	positions := c.Request().Form["choice_position[]"]
	var choices []services.FieldChoiceInput
	for i, label := range labels {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		pos := i + 1
		if i < len(positions) {
			if p, err := strconv.Atoi(positions[i]); err == nil {
				pos = p
			}
		}
		choices = append(choices, services.FieldChoiceInput{Label: label, Position: pos})
	}
	return choices
}

func jsonError(c buffalo.Context, status int, msg string) error {
	return c.Render(status, r.JSON(map[string]string{"error": msg}))
}
