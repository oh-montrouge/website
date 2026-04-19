package grifts

import (
	"fmt"

	"ohmontrouge/webapp/models"
	"ohmontrouge/webapp/services"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/grift/grift"
)

func accountService() services.AccountService {
	return services.AccountService{
		Accounts: models.AccountStore{},
		Roles:    models.AccountRoleStore{},
	}
}

var _ = grift.Namespace("db", func() {
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
