package models

import "github.com/gobuffalo/pop/v6"

// Instrument represents a musical instrument from the controlled list.
type Instrument struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

// Instruments is a slice of Instrument.
type Instruments []Instrument

// InstrumentsList returns all instruments ordered by name.
func InstrumentsList(tx *pop.Connection) (Instruments, error) {
	instruments := Instruments{}
	err := tx.Order("name ASC").All(&instruments)
	return instruments, err
}

// InstrumentStore is the production implementation of actions.InstrumentRepository.
type InstrumentStore struct{}

func (InstrumentStore) List(tx *pop.Connection) (Instruments, error) {
	return InstrumentsList(tx)
}
