package services_test

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/argon2"
	"ohmontrouge/webapp/models"
	"ohmontrouge/webapp/services"
)

// stubAccountRepo is a stub AccountRepository for unit tests.
type stubAccountRepo struct {
	account   *models.Account
	err       error
	createdID int64
}

func (s stubAccountRepo) FindByEmail(_ *pop.Connection, _ string) (*models.Account, error) {
	return s.account, s.err
}

func (s stubAccountRepo) GetByID(_ *pop.Connection, _ int64) (*models.Account, error) {
	return s.account, s.err
}

func (s stubAccountRepo) Create(_ *pop.Connection, _, _ string, _ int64) (int64, error) {
	return s.createdID, s.err
}

func (s stubAccountRepo) UpdatePasswordHash(_ *pop.Connection, _ int64, _ string) error {
	return s.err
}

// stubRoleRepo is a stub RoleRepository for unit tests.
type stubRoleRepo struct {
	hasActiveHolder bool
	holderErr       error
	roleID          int64
	roleIDErr       error
	assignErr       error
}

func (s stubRoleRepo) HasRole(_ *pop.Connection, _ int64, _ string) (bool, error) {
	return false, nil
}

func (s stubRoleRepo) HasActiveRoleHolder(_ *pop.Connection, _ string) (bool, error) {
	return s.hasActiveHolder, s.holderErr
}

func (s stubRoleRepo) GetIDByName(_ *pop.Connection, _ string) (int64, error) {
	return s.roleID, s.roleIDErr
}

func (s stubRoleRepo) AssignRole(_ *pop.Connection, _, _ int64) error {
	return s.assignErr
}

// testArgon2Hash generates a PHC-format argon2id hash with minimal parameters for test speed.
func testArgon2Hash(t *testing.T, password string) string {
	t.Helper()
	salt := []byte("testsalt12345678") // 16 bytes, fixed for reproducibility
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 1, 32)
	return fmt.Sprintf("$argon2id$v=%d$m=65536,t=1,p=1$%s$%s",
		argon2.Version,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
}

func TestAuthenticate_AccountNotFound(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{err: sql.ErrNoRows},
	}
	_, err := svc.Authenticate(nil, "nobody@example.com", "password")
	assert.ErrorIs(t, err, services.ErrAccountNotFound)
}

func TestAuthenticate_DBError(t *testing.T) {
	dbErr := errors.New("connection refused")
	svc := services.AccountService{
		Accounts: stubAccountRepo{err: dbErr},
	}
	_, err := svc.Authenticate(nil, "alice@example.com", "password")
	assert.ErrorIs(t, err, dbErr)
}

func TestAuthenticate_InactiveAccount(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{account: &models.Account{
			Email:        nulls.NewString("pending@example.com"),
			PasswordHash: nulls.NewString(testArgon2Hash(t, "correct")),
			Status:       "pending",
		}},
	}
	_, err := svc.Authenticate(nil, "pending@example.com", "correct")
	assert.ErrorIs(t, err, services.ErrInvalidPassword)
}

func TestAuthenticate_NullPasswordHash(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{account: &models.Account{
			Email:        nulls.NewString("active@example.com"),
			PasswordHash: nulls.String{}, // active but not yet activated (defense-in-depth)
			Status:       "active",
		}},
	}
	_, err := svc.Authenticate(nil, "active@example.com", "password")
	assert.ErrorIs(t, err, services.ErrInvalidPassword)
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{account: &models.Account{
			Email:        nulls.NewString("alice@example.com"),
			PasswordHash: nulls.NewString(testArgon2Hash(t, "correct")),
			Status:       "active",
		}},
	}
	_, err := svc.Authenticate(nil, "alice@example.com", "wrong")
	assert.ErrorIs(t, err, services.ErrInvalidPassword)
}

func TestAuthenticate_Success(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{account: &models.Account{
			ID:           1,
			Email:        nulls.NewString("alice@example.com"),
			PasswordHash: nulls.NewString(testArgon2Hash(t, "correct")),
			Status:       "active",
		}},
	}
	got, err := svc.Authenticate(nil, "alice@example.com", "correct")
	assert.NoError(t, err)
	assert.Equal(t, &services.AccountDTO{
		ID:     1,
		Email:  "alice@example.com",
		Status: services.StatusActive,
	}, got)
}

func TestCreateAdmin_Success(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{createdID: 1},
		Roles:    stubRoleRepo{roleID: 1},
	}
	err := svc.CreateAdmin(nil, "admin@example.com", "secret", 1)
	assert.NoError(t, err)
}

func TestCreateAdmin_AlreadyExists(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{},
		Roles:    stubRoleRepo{hasActiveHolder: true},
	}
	err := svc.CreateAdmin(nil, "admin@example.com", "secret", 1)
	assert.ErrorIs(t, err, services.ErrAdminAlreadyExists)
}

func TestResetPassword_Success(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{account: &models.Account{
			ID:     1,
			Email:  nulls.NewString("admin@example.com"),
			Status: "active",
		}},
		Roles: stubRoleRepo{},
	}
	err := svc.ResetPassword(nil, "admin@example.com", "newpassword")
	assert.NoError(t, err)
}

func TestResetPassword_AccountNotFound(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{err: sql.ErrNoRows},
		Roles:    stubRoleRepo{},
	}
	err := svc.ResetPassword(nil, "nobody@example.com", "newpassword")
	assert.ErrorIs(t, err, services.ErrAccountNotFound)
}

func TestResetPassword_AccountNotActive(t *testing.T) {
	svc := services.AccountService{
		Accounts: stubAccountRepo{account: &models.Account{
			ID:     1,
			Email:  nulls.NewString("pending@example.com"),
			Status: "pending",
		}},
		Roles: stubRoleRepo{},
	}
	err := svc.ResetPassword(nil, "pending@example.com", "newpassword")
	assert.ErrorIs(t, err, services.ErrAccountNotActive)
}
