package models

import "testing"

func TestInstrumentsList_ReturnsSortedByName(t *testing.T) {
	truncateAll(t)

	// Insert out-of-order to verify the ORDER BY name ASC clause.
	if err := DB.Create(&Instrument{Name: "Violon"}); err != nil {
		t.Fatal(err)
	}
	if err := DB.Create(&Instrument{Name: "Alto"}); err != nil {
		t.Fatal(err)
	}
	if err := DB.Create(&Instrument{Name: "Cor"}); err != nil {
		t.Fatal(err)
	}

	instruments, err := InstrumentsList(DB)
	if err != nil {
		t.Fatal(err)
	}

	if len(instruments) != 3 {
		t.Fatalf("expected 3 instruments, got %d", len(instruments))
	}
	if instruments[0].Name != "Alto" || instruments[1].Name != "Cor" || instruments[2].Name != "Violon" {
		t.Errorf("unexpected order: %v", instruments)
	}
}
