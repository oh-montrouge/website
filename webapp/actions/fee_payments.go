package actions

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"ohmontrouge/webapp/services"
)

// FeePaymentsHandler handles admin fee payment CRUD routes.
type FeePaymentsHandler struct {
	FeePayments services.FeePaymentManager
}

var validPaymentTypes = map[string]bool{
	"chèque":            true,
	"espèces":           true,
	"virement bancaire": true,
}

// Create records a new fee payment for the account identified by account_id.
func (h FeePaymentsHandler) Create(c buffalo.Context) error {
	accountID, err := strconv.ParseInt(c.Param("account_id"), 10, 64)
	if err != nil {
		return c.Error(http.StatusBadRequest, err)
	}

	seasonIDStr := c.Request().FormValue("season_id")
	amountStr := strings.TrimSpace(c.Request().FormValue("amount"))
	paymentDateStr := c.Request().FormValue("payment_date")
	paymentType := strings.TrimSpace(c.Request().FormValue("payment_type"))
	comment := strings.TrimSpace(c.Request().FormValue("comment"))

	seasonID, err := strconv.ParseInt(seasonIDStr, 10, 64)
	if err != nil || seasonID <= 0 {
		c.Flash().Add("danger", "Saison invalide.")
		return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", accountID)
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount < 0 {
		c.Flash().Add("danger", "Montant invalide.")
		return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", accountID)
	}

	paymentDate, err := time.Parse("2006-01-02", paymentDateStr)
	if err != nil {
		c.Flash().Add("danger", "Date de paiement invalide.")
		return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", accountID)
	}

	if !validPaymentTypes[paymentType] {
		c.Flash().Add("danger", "Le type de paiement est invalide.")
		return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", accountID)
	}

	tx := c.Value("tx").(*pop.Connection)
	if err := h.FeePayments.Record(tx, accountID, seasonID, amount, paymentDate, paymentType, comment); err != nil {
		if errors.Is(err, services.ErrDuplicatePayment) {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", accountID)
		}
		return err
	}

	c.Flash().Add("success", "Cotisation enregistrée.")
	return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", accountID)
}

// loadFeePayment parses the :id param, fetches the payment, and returns 404 on failure.
func loadFeePayment(c buffalo.Context, fp services.FeePaymentManager) (*services.FeePaymentDTO, *pop.Connection, error) {
	id, err := parseID(c)
	if err != nil {
		return nil, nil, c.Error(http.StatusNotFound, err)
	}
	tx := c.Value("tx").(*pop.Connection)
	payment, err := fp.GetByID(tx, id)
	if err != nil {
		if errors.Is(err, services.ErrFeePaymentNotFound) {
			return nil, nil, c.Error(http.StatusNotFound, err)
		}
		return nil, nil, err
	}
	return payment, tx, nil
}

// EditForm renders the fee payment edit form.
func (h FeePaymentsHandler) EditForm(c buffalo.Context) error {
	payment, _, err := loadFeePayment(c, h.FeePayments)
	if err != nil {
		return err
	}
	c.Set("payment", payment)
	c.Set("formError", "")
	return c.Render(http.StatusOK, r.HTML("admin/cotisations/edit.plush.html"))
}

// Update saves edits to a fee payment.
func (h FeePaymentsHandler) Update(c buffalo.Context) error {
	payment, tx, err := loadFeePayment(c, h.FeePayments)
	if err != nil {
		return err
	}

	amountStr := strings.TrimSpace(c.Request().FormValue("amount"))
	paymentDateStr := c.Request().FormValue("payment_date")
	paymentType := strings.TrimSpace(c.Request().FormValue("payment_type"))
	comment := strings.TrimSpace(c.Request().FormValue("comment"))

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount < 0 {
		c.Set("payment", payment)
		c.Set("formError", "Montant invalide.")
		return c.Render(http.StatusUnprocessableEntity, r.HTML("admin/cotisations/edit.plush.html"))
	}

	paymentDate, err := time.Parse("2006-01-02", paymentDateStr)
	if err != nil {
		c.Set("payment", payment)
		c.Set("formError", "Date de paiement invalide.")
		return c.Render(http.StatusUnprocessableEntity, r.HTML("admin/cotisations/edit.plush.html"))
	}

	if !validPaymentTypes[paymentType] {
		c.Set("payment", payment)
		c.Set("formError", "Le type de paiement est invalide.")
		return c.Render(http.StatusUnprocessableEntity, r.HTML("admin/cotisations/edit.plush.html"))
	}

	if err := h.FeePayments.Update(tx, payment.ID, amount, paymentDate, paymentType, comment); err != nil {
		return err
	}

	c.Flash().Add("success", "Cotisation mise à jour.")
	return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", payment.AccountID)
}

// Delete removes a fee payment.
func (h FeePaymentsHandler) Delete(c buffalo.Context) error {
	payment, tx, err := loadFeePayment(c, h.FeePayments)
	if err != nil {
		return err
	}

	if err := h.FeePayments.Delete(tx, payment.ID); err != nil {
		return err
	}

	c.Flash().Add("success", "Cotisation supprimée.")
	return c.Redirect(http.StatusSeeOther, "/admin/musiciens/%d", payment.AccountID)
}
