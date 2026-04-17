package models

import "github.com/gobuffalo/pop/v6"

// Role represents an application role (e.g. "admin").
type Role struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

// AccountRole is the join between an account and a role.
type AccountRole struct {
	AccountID int64 `db:"account_id"`
	RoleID    int64 `db:"role_id"`
}

// AccountRoleStore is the production implementation of services.RoleRepository.
type AccountRoleStore struct{}

type countResult struct {
	N int `db:"count"`
}

func (AccountRoleStore) HasRole(tx *pop.Connection, accountID int64, roleName string) (bool, error) {
	var res countResult
	err := tx.RawQuery(
		`SELECT COUNT(*) AS count FROM account_roles
		 JOIN roles ON roles.id = account_roles.role_id
		 WHERE account_roles.account_id = ? AND roles.name = ?`,
		accountID, roleName,
	).First(&res)
	return res.N > 0, err
}

func (AccountRoleStore) HasActiveRoleHolder(tx *pop.Connection, roleName string) (bool, error) {
	var res countResult
	err := tx.RawQuery(
		`SELECT COUNT(*) AS count
		 FROM accounts
		 JOIN account_roles ON account_roles.account_id = accounts.id
		 JOIN roles         ON roles.id = account_roles.role_id
		 WHERE accounts.status = 'active' AND roles.name = ?`,
		roleName,
	).First(&res)
	return res.N > 0, err
}

func (AccountRoleStore) GetIDByName(tx *pop.Connection, name string) (int64, error) {
	var row struct {
		ID int64 `db:"id"`
	}
	err := tx.RawQuery(`SELECT id FROM roles WHERE name = ?`, name).First(&row)
	return row.ID, err
}

func (AccountRoleStore) AssignRole(tx *pop.Connection, accountID, roleID int64) error {
	return tx.RawQuery(
		`INSERT INTO account_roles (account_id, role_id) VALUES (?, ?)`,
		accountID, roleID,
	).Exec()
}
