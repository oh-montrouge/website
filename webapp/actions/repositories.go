package actions

import (
	"github.com/gobuffalo/pop/v6"
	"ohmontrouge/webapp/models"
)

// InstrumentRepository is the interface handlers depend on to access instrument data.
// Defined on the consumer side (actions) per the Dependency Inversion Principle.
// The real implementation is models.InstrumentStore; tests inject stubs.
type InstrumentRepository interface {
	List(tx *pop.Connection) (models.Instruments, error)
}
