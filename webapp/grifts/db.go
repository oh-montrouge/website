package grifts

import (
	"fmt"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/grift/grift"
	"github.com/gobuffalo/pop/v6"
	"ohmontrouge/webapp/models"
	"ohmontrouge/webapp/services"
)

func accountService() services.AccountService {
	return services.AccountService{
		Accounts: models.AccountStore{},
		Roles:    models.AccountRoleStore{},
	}
}

var _ = grift.Namespace("db", func() {

	_ = grift.Desc("seed:admin", "Creates the first active admin account (ADMIN_EMAIL, ADMIN_PASSWORD)")
	_ = grift.Add("seed:admin", func(c *grift.Context) error {
		email, err := envy.MustGet("ADMIN_EMAIL")
		if err != nil {
			return fmt.Errorf("ADMIN_EMAIL not set: %w", err)
		}
		password, err := envy.MustGet("ADMIN_PASSWORD")
		if err != nil {
			return fmt.Errorf("ADMIN_PASSWORD not set: %w", err)
		}

		var instrument struct {
			ID int64 `db:"id"`
		}
		if err := models.DB.RawQuery(
			`SELECT id FROM instruments WHERE name = ? LIMIT 1`, "Chef d'orchestre",
		).First(&instrument); err != nil {
			if err2 := models.DB.RawQuery(
				`SELECT id FROM instruments ORDER BY id LIMIT 1`,
			).First(&instrument); err2 != nil {
				return fmt.Errorf("no instruments found: %w", err2)
			}
		}

		svc := accountService()
		return models.DB.Transaction(func(tx *pop.Connection) error {
			if err := svc.CreateAdmin(tx, email, password, instrument.ID); err != nil {
				return err
			}
			fmt.Printf("Admin account created: %s\n", email)
			return nil
		})
	})

	_ = grift.Desc("recover:admin", "Force-resets password for an active account (ADMIN_EMAIL, ADMIN_NEW_PASSWORD)")
	_ = grift.Add("recover:admin", func(c *grift.Context) error {
		email, err := envy.MustGet("ADMIN_EMAIL")
		if err != nil {
			return fmt.Errorf("ADMIN_EMAIL not set: %w", err)
		}
		newPassword, err := envy.MustGet("ADMIN_NEW_PASSWORD")
		if err != nil {
			return fmt.Errorf("ADMIN_NEW_PASSWORD not set: %w", err)
		}

		svc := accountService()
		if err := svc.ResetPassword(models.DB, email, newPassword); err != nil {
			return err
		}
		fmt.Printf("Password reset for: %s\n", email)
		return nil
	})

})
