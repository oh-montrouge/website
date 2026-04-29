package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	s24Start = time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)
	s24End   = time.Date(2025, 8, 31, 0, 0, 0, 0, time.UTC)
	s25Start = time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	s25End   = time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
)

func TestSeasonStore_Create(t *testing.T) {
	truncateAll(t)
	store := SeasonStore{}

	id, err := store.Create(DB, "2025-2026", s25Start, s25End)
	require.NoError(t, err)
	assert.Positive(t, id)

	var s Season
	require.NoError(t, DB.Find(&s, id))
	assert.Equal(t, "2025-2026", s.Label)
	assert.False(t, s.IsCurrent, "AC-M1: new season must not be current")
}

func TestSeasonStore_List_Empty(t *testing.T) {
	truncateAll(t)
	store := SeasonStore{}

	ss, err := store.List(DB)
	require.NoError(t, err)
	assert.Empty(t, ss)
}

func TestSeasonStore_List_OrderedByStartDateDesc(t *testing.T) {
	truncateAll(t)
	store := SeasonStore{}

	// Insert older season first, newer second — List must return newest first.
	older := &Season{Label: "2024-2025", StartDate: s24Start, EndDate: s24End}
	newer := &Season{Label: "2025-2026", StartDate: s25Start, EndDate: s25End}
	require.NoError(t, DB.Create(older))
	require.NoError(t, DB.Create(newer))

	ss, err := store.List(DB)
	require.NoError(t, err)
	require.Len(t, ss, 2)
	assert.Equal(t, "2025-2026", ss[0].Label, "most recent season first")
	assert.Equal(t, "2024-2025", ss[1].Label)
}

// seedTwoSeasons truncates, creates two seasons (2024-2025 current, 2025-2026 next), and returns them.
func seedTwoSeasons(t *testing.T) (SeasonStore, *Season, *Season) {
	t.Helper()
	truncateAll(t)
	store := SeasonStore{}
	s1 := &Season{Label: "2024-2025", StartDate: s24Start, EndDate: s24End, IsCurrent: true}
	s2 := &Season{Label: "2025-2026", StartDate: s25Start, EndDate: s25End}
	require.NoError(t, DB.Create(s1))
	require.NoError(t, DB.Create(s2))
	return store, s1, s2
}

// TestSeasonStore_DesignateCurrent_TransfersDesignation verifies that calling
// DesignateCurrent clears the previous current and sets the new one (AC-M2).
func TestSeasonStore_DesignateCurrent_TransfersDesignation(t *testing.T) {
	store, current, next := seedTwoSeasons(t)

	require.NoError(t, store.DesignateCurrent(DB, next.ID))

	var gotCurrent, gotNext Season
	require.NoError(t, DB.Find(&gotCurrent, current.ID))
	require.NoError(t, DB.Find(&gotNext, next.ID))

	assert.False(t, gotCurrent.IsCurrent, "previous current must be cleared")
	assert.True(t, gotNext.IsCurrent)
}

// TestSeasonStore_DesignateCurrent_ExactlyOneCurrent verifies the exactly-one invariant (AC-M2).
func TestSeasonStore_DesignateCurrent_ExactlyOneCurrent(t *testing.T) {
	store, _, s2 := seedTwoSeasons(t)

	require.NoError(t, store.DesignateCurrent(DB, s2.ID))

	var result struct {
		Count int `db:"count"`
	}
	require.NoError(t, DB.RawQuery("SELECT COUNT(*) AS count FROM seasons WHERE is_current = true").First(&result))
	assert.Equal(t, 1, result.Count)
}

// TestSeasonStore_DesignateCurrent_Idempotent verifies that multiple sequential calls
// leave exactly one current season at the end (AC-M3).
func TestSeasonStore_DesignateCurrent_Idempotent(t *testing.T) {
	store, s1, s2 := seedTwoSeasons(t)

	require.NoError(t, store.DesignateCurrent(DB, s2.ID))
	require.NoError(t, store.DesignateCurrent(DB, s1.ID))
	require.NoError(t, store.DesignateCurrent(DB, s2.ID))

	var got1, got2 Season
	require.NoError(t, DB.Find(&got1, s1.ID))
	require.NoError(t, DB.Find(&got2, s2.ID))
	assert.False(t, got1.IsCurrent)
	assert.True(t, got2.IsCurrent)

	var result struct {
		Count int `db:"count"`
	}
	require.NoError(t, DB.RawQuery("SELECT COUNT(*) AS count FROM seasons WHERE is_current = true").First(&result))
	assert.Equal(t, 1, result.Count)
}
