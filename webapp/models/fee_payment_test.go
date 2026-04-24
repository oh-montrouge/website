package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	fpDate1 = time.Date(2024, 9, 15, 0, 0, 0, 0, time.UTC)
	fpDate2 = time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC)
	fpDate3 = time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC)
)

func insertTestAccountAndSeason(t *testing.T) (accountID, seasonID int64) {
	t.Helper()

	var instrRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(
		`INSERT INTO instruments (name) VALUES ('TestInstrument') RETURNING id`,
	).First(&instrRow))

	var accRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(
		`INSERT INTO accounts (email, password_hash, main_instrument_id, status)
		 VALUES ('fp_test@example.com', 'hash', ?, 'active') RETURNING id`,
		instrRow.ID,
	).First(&accRow))

	var seasonRow struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(
		`INSERT INTO seasons (label, start_date, end_date, is_current) VALUES ('2025-2026', '2025-09-01', '2026-08-31', false) RETURNING id`,
	).First(&seasonRow))

	return accRow.ID, seasonRow.ID
}

func TestFeePaymentStore_Create_Success(t *testing.T) {
	truncateAll(t)
	store := FeePaymentStore{}
	accountID, seasonID := insertTestAccountAndSeason(t)

	id, err := store.Create(DB, accountID, seasonID, 50.00, fpDate1, "chèque", "")
	require.NoError(t, err)
	assert.Positive(t, id)

	var fp FeePayment
	require.NoError(t, DB.Find(&fp, id))
	assert.Equal(t, accountID, fp.AccountID)
	assert.Equal(t, seasonID, fp.SeasonID)
	assert.InDelta(t, 50.00, fp.Amount, 0.001)
	assert.Equal(t, "chèque", fp.PaymentType)
	assert.False(t, fp.Comment.Valid, "empty comment should be stored as NULL")
}

// TestFeePaymentStore_Create_Duplicate covers AC-M1: duplicate (account, season) is rejected.
func TestFeePaymentStore_Create_Duplicate(t *testing.T) {
	truncateAll(t)
	store := FeePaymentStore{}
	accountID, seasonID := insertTestAccountAndSeason(t)

	_, err := store.Create(DB, accountID, seasonID, 50.00, fpDate1, "chèque", "")
	require.NoError(t, err)

	_, err = store.Create(DB, accountID, seasonID, 60.00, fpDate2, "espèces", "")
	assert.ErrorIs(t, err, ErrFeePaymentDuplicate, "AC-M1: second payment for same (account, season) must be rejected")

	var result struct {
		Count int `db:"count"`
	}
	require.NoError(t, DB.RawQuery(
		`SELECT COUNT(*) AS count FROM fee_payments WHERE account_id=? AND season_id=?`,
		accountID, seasonID,
	).First(&result))
	assert.Equal(t, 1, result.Count, "AC-M1: only one payment row must exist")
}

func TestFeePaymentStore_Create_WithComment(t *testing.T) {
	truncateAll(t)
	store := FeePaymentStore{}
	accountID, seasonID := insertTestAccountAndSeason(t)

	id, err := store.Create(DB, accountID, seasonID, 50.00, fpDate1, "espèces", "test note")
	require.NoError(t, err)

	var fp FeePayment
	require.NoError(t, DB.Find(&fp, id))
	assert.True(t, fp.Comment.Valid)
	assert.Equal(t, "test note", fp.Comment.String)
}

func TestFeePaymentStore_ListByAccount_Empty(t *testing.T) {
	truncateAll(t)
	store := FeePaymentStore{}
	accountID, _ := insertTestAccountAndSeason(t)

	rows, err := store.ListByAccount(DB, accountID)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestFeePaymentStore_ListByAccount_ReturnsPaymentsWithSeasonLabel(t *testing.T) {
	truncateAll(t)
	store := FeePaymentStore{}
	accountID, seasonID := insertTestAccountAndSeason(t)

	_, err := store.Create(DB, accountID, seasonID, 50.00, fpDate1, "chèque", "")
	require.NoError(t, err)

	rows, err := store.ListByAccount(DB, accountID)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "2025-2026", rows[0].SeasonLabel)
	assert.InDelta(t, 50.00, rows[0].Amount, 0.001)
}

func TestFeePaymentStore_GetByID_Found(t *testing.T) {
	truncateAll(t)
	store := FeePaymentStore{}
	accountID, seasonID := insertTestAccountAndSeason(t)

	id, err := store.Create(DB, accountID, seasonID, 50.00, fpDate1, "virement bancaire", "note")
	require.NoError(t, err)

	row, err := store.GetByID(DB, id)
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, id, row.ID)
	assert.Equal(t, "2025-2026", row.SeasonLabel)
	assert.Equal(t, "virement bancaire", row.PaymentType)
}

func TestFeePaymentStore_GetByID_NotFound(t *testing.T) {
	truncateAll(t)
	store := FeePaymentStore{}

	row, err := store.GetByID(DB, 9999)
	require.NoError(t, err)
	assert.Nil(t, row)
}

func TestFeePaymentStore_Update(t *testing.T) {
	truncateAll(t)
	store := FeePaymentStore{}
	accountID, seasonID := insertTestAccountAndSeason(t)

	id, err := store.Create(DB, accountID, seasonID, 50.00, fpDate1, "chèque", "")
	require.NoError(t, err)

	require.NoError(t, store.Update(DB, id, 75.00, fpDate2, "espèces", "updated"))

	row, err := store.GetByID(DB, id)
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.InDelta(t, 75.00, row.Amount, 0.001)
	assert.Equal(t, "espèces", row.PaymentType)
	assert.Equal(t, "updated", row.Comment.String)
}

// TestFeePaymentStore_Delete covers AC-M3: delete removes only the target row.
func TestFeePaymentStore_Delete(t *testing.T) {
	truncateAll(t)
	store := FeePaymentStore{}
	accountID, seasonID := insertTestAccountAndSeason(t)

	// Insert a second season for the second payment.
	var s2Row struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(
		`INSERT INTO seasons (label, start_date, end_date, is_current) VALUES ('2024-2025', '2024-09-01', '2025-08-31', false) RETURNING id`,
	).First(&s2Row))
	season2ID := s2Row.ID

	id1, err := store.Create(DB, accountID, seasonID, 50.00, fpDate1, "chèque", "")
	require.NoError(t, err)
	id2, err := store.Create(DB, accountID, season2ID, 45.00, fpDate3, "espèces", "")
	require.NoError(t, err)

	require.NoError(t, store.Delete(DB, id1))

	row1, err := store.GetByID(DB, id1)
	require.NoError(t, err)
	assert.Nil(t, row1, "AC-M3: deleted payment must not exist")

	row2, err := store.GetByID(DB, id2)
	require.NoError(t, err)
	assert.NotNil(t, row2, "AC-M3: other payment must remain")
}

// TestFeePaymentStore_GetFirstInscriptionDate_WithPayments covers AC-M2.
func TestFeePaymentStore_GetFirstInscriptionDate_WithPayments(t *testing.T) {
	truncateAll(t)
	store := FeePaymentStore{}
	accountID, seasonID := insertTestAccountAndSeason(t)

	var s2Row struct {
		ID int64 `db:"id"`
	}
	require.NoError(t, DB.RawQuery(
		`INSERT INTO seasons (label, start_date, end_date, is_current) VALUES ('2024-2025', '2024-09-01', '2025-08-31', false) RETURNING id`,
	).First(&s2Row))
	season2ID := s2Row.ID

	// Three payments on different dates; earliest is fpDate3 (2024-11-01 < fpDate1 2024-09-15... wait)
	// fpDate1 = 2024-09-15, fpDate3 = 2024-11-01, fpDate2 = 2025-09-10
	// Earliest is fpDate1.
	_, err := store.Create(DB, accountID, seasonID, 50.00, fpDate2, "chèque", "")
	require.NoError(t, err)
	_, err = store.Create(DB, accountID, season2ID, 50.00, fpDate1, "espèces", "")
	require.NoError(t, err)

	date, err := store.GetFirstInscriptionDate(DB, accountID)
	require.NoError(t, err)
	require.NotNil(t, date, "AC-M2: account with payments must have a first inscription date")
	assert.Equal(t, fpDate1.Format("2006-01-02"), date.Format("2006-01-02"),
		"AC-M2: first inscription date must be the earliest payment date")
}

// TestFeePaymentStore_GetFirstInscriptionDate_NoPayments covers AC-M2 nil case.
func TestFeePaymentStore_GetFirstInscriptionDate_NoPayments(t *testing.T) {
	truncateAll(t)
	store := FeePaymentStore{}
	accountID, _ := insertTestAccountAndSeason(t)

	date, err := store.GetFirstInscriptionDate(DB, accountID)
	require.NoError(t, err)
	assert.Nil(t, date, "AC-M2: account with no payments must return nil")
}
