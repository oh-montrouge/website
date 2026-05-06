package services

import (
	"testing"

	"ohmontrouge/webapp/models"

	"github.com/gobuffalo/nulls"
	"github.com/stretchr/testify/assert"
)

// --- toSummaryDTOs ---

func TestToSummaryDTOs_NullRSVPState(t *testing.T) {
	rows := []models.EventListRow{{ID: 1, Name: "Concert", EventType: "concert"}}
	dtos := toSummaryDTOs(rows)
	assert.Equal(t, "", dtos[0].RSVPState)
}

func TestToSummaryDTOs_ValidRSVPState(t *testing.T) {
	rows := []models.EventListRow{
		{ID: 1, Name: "Concert", EventType: "concert", RSVPState: nulls.NewString("yes")},
	}
	dtos := toSummaryDTOs(rows)
	assert.Equal(t, "yes", dtos[0].RSVPState)
}

// --- displayName ---

func TestDisplayName_FullName(t *testing.T) {
	r := models.RSVPListRow{
		FirstName: nulls.NewString("Alice"),
		LastName:  nulls.NewString("Dupont"),
	}
	assert.Equal(t, "Dupont Alice", displayName(r))
}

func TestDisplayName_Anonymized(t *testing.T) {
	r := models.RSVPListRow{
		AnonymizationToken: nulls.NewString("abcdef1234567890"),
	}
	assert.Equal(t, "Musicien abcdef12", displayName(r))
}

func TestDisplayName_Unknown(t *testing.T) {
	r := models.RSVPListRow{}
	assert.Equal(t, "Compte inconnu", displayName(r))
}

// --- buildPupitre ---

func TestBuildPupitre_Empty(t *testing.T) {
	assert.Empty(t, buildPupitre(nil))
}

func TestBuildPupitre_UsesRSVPInstrumentForYes(t *testing.T) {
	rows := []models.RSVPListRow{
		{State: "yes", RSVPInstrumentName: nulls.NewString("Violon"), MainInstrumentName: "Alto"},
	}
	result := buildPupitre(rows)
	assert.Len(t, result, 1)
	assert.Equal(t, "Violon", result[0].InstrumentName)
	assert.Equal(t, 1, result[0].Yes)
}

func TestBuildPupitre_UsesMainInstrumentForNonYes(t *testing.T) {
	rows := []models.RSVPListRow{
		{State: "no", RSVPInstrumentName: nulls.NewString("Violon"), MainInstrumentName: "Alto"},
		{State: "maybe", RSVPInstrumentName: nulls.NewString("Violon"), MainInstrumentName: "Alto"},
		{State: "unanswered", RSVPInstrumentName: nulls.NewString("Violon"), MainInstrumentName: "Alto"},
	}
	result := buildPupitre(rows)
	assert.Len(t, result, 1)
	assert.Equal(t, "Alto", result[0].InstrumentName)
	assert.Equal(t, 1, result[0].No)
	assert.Equal(t, 1, result[0].Maybe)
	assert.Equal(t, 1, result[0].Unanswered)
}

func TestBuildPupitre_AllStates(t *testing.T) {
	rows := []models.RSVPListRow{
		{State: "yes", RSVPInstrumentName: nulls.NewString("Violon"), MainInstrumentName: "Violon"},
		{State: "maybe", MainInstrumentName: "Violon"},
		{State: "no", MainInstrumentName: "Violon"},
		{State: "unanswered", MainInstrumentName: "Violon"},
	}
	result := buildPupitre(rows)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, result[0].Yes)
	assert.Equal(t, 1, result[0].Maybe)
	assert.Equal(t, 1, result[0].No)
	assert.Equal(t, 1, result[0].Unanswered)
}

func TestBuildPupitre_PreservesOrder(t *testing.T) {
	rows := []models.RSVPListRow{
		{State: "yes", RSVPInstrumentName: nulls.NewString("Violon"), MainInstrumentName: "Violon"},
		{State: "yes", RSVPInstrumentName: nulls.NewString("Alto"), MainInstrumentName: "Alto"},
		{State: "yes", RSVPInstrumentName: nulls.NewString("Violon"), MainInstrumentName: "Violon"},
	}
	result := buildPupitre(rows)
	assert.Len(t, result, 2)
	assert.Equal(t, "Violon", result[0].InstrumentName)
	assert.Equal(t, "Alto", result[1].InstrumentName)
}
