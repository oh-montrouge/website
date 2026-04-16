package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
)

// InstrumentsHandler handles routes related to the instruments list.
type InstrumentsHandler struct {
	Instruments InstrumentRepository
}

// Index lists all instruments, ordered by name.
func (h InstrumentsHandler) Index(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	instruments, err := h.Instruments.List(tx)
	if err != nil {
		return err
	}
	c.Set("instruments", instruments)
	return c.Render(200, r.HTML("instruments/index.plush.html"))
}
