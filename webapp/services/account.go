package services

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/gobuffalo/pop/v6"
	"golang.org/x/crypto/argon2"
	"ohmontrouge/webapp/models"
)

// AccountStatus represents the lifecycle state of a member account.
type AccountStatus string

const (
	StatusPending    AccountStatus = "pending"
	StatusActive     AccountStatus = "active"
	StatusAnonymized AccountStatus = "anonymized"
)

// RoleAdmin is the name of the administrator role as stored in the roles table.
const RoleAdmin = "admin"

var (
	ErrAccountNotFound    = errors.New("account not found")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrAdminAlreadyExists = errors.New("an active admin account already exists")
	ErrAccountNotActive   = errors.New("account is not active")
)

// AccountAuthenticator is the interface auth handlers and middleware depend on.
type AccountAuthenticator interface {
	Authenticate(tx *pop.Connection, email, password string) (*AccountDTO, error)
	GetByID(tx *pop.Connection, id int64) (*AccountDTO, error)
	IsAdmin(tx *pop.Connection, accountID int64) (bool, error)
}

// AccountDTO is the Identity & Access view of an account, used by auth handlers and
// session middleware. It carries only the fields needed to establish and validate a session.
// Membership fields (name, instrument, profile data) are absent by design — see
// specs/technical-adrs/007-account-musician-dtos.md.
// Sensitive fields (PasswordHash, AnonymizationToken) are absent by construction.
type AccountDTO struct {
	ID     int64
	Email  string
	Status AccountStatus
}

// AccountService implements domain logic for account operations.
type AccountService struct {
	Accounts AccountRepository
	Roles    RoleRepository
}

// IsAdmin reports whether the account with the given ID holds the admin role.
func (s AccountService) IsAdmin(tx *pop.Connection, accountID int64) (bool, error) {
	return s.Roles.HasRole(tx, accountID, RoleAdmin)
}

// CreateAdmin creates an active account with the admin role.
// Returns ErrAdminAlreadyExists if an active admin already exists.
func (s AccountService) CreateAdmin(tx *pop.Connection, email, password string, instrumentID int64) error {
	exists, err := s.Roles.HasActiveRoleHolder(tx, RoleAdmin)
	if err != nil {
		return err
	}
	if exists {
		return ErrAdminAlreadyExists
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	accountID, err := s.Accounts.Create(tx, email, hash, instrumentID)
	if err != nil {
		return err
	}
	roleID, err := s.Roles.GetIDByName(tx, RoleAdmin)
	if err != nil {
		return err
	}
	return s.Roles.AssignRole(tx, accountID, roleID)
}

// ResetPassword force-resets the password for the active account with the given email.
// Returns ErrAccountNotFound or ErrAccountNotActive if the account cannot be found or is not active.
func (s AccountService) ResetPassword(tx *pop.Connection, email, newPassword string) error {
	account, err := s.Accounts.FindByEmail(tx, email)
	if err != nil {
		if isNotFound(err) {
			return ErrAccountNotFound
		}
		return err
	}
	if AccountStatus(account.Status) != StatusActive {
		return ErrAccountNotActive
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.Accounts.UpdatePasswordHash(tx, account.ID, hash)
}

// HashPassword generates an argon2id PHC-format hash for a plaintext password.
// Parameters follow OWASP minimums: t=3, m=64MB, p=4, key=32 bytes.
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
	return fmt.Sprintf("$argon2id$v=%d$m=65536,t=3,p=4$%s$%s",
		argon2.Version,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// Authenticate verifies email/password and returns the matching account.
// Both "not found" and "wrong password" surface as distinct sentinel errors so
// callers can log them separately, but must present them identically to the
// user to prevent account enumeration.
func (s AccountService) Authenticate(tx *pop.Connection, email, password string) (*AccountDTO, error) {
	account, err := s.Accounts.FindByEmail(tx, email)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	if AccountStatus(account.Status) != StatusActive {
		return nil, ErrInvalidPassword
	}
	if !account.PasswordHash.Valid {
		return nil, ErrInvalidPassword
	}
	ok, err := verifyArgon2id(password, account.PasswordHash.String)
	if err != nil || !ok {
		return nil, ErrInvalidPassword
	}
	return toAccountDTO(account), nil
}

func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

// GetByID returns the account with the given ID.
func (s AccountService) GetByID(tx *pop.Connection, id int64) (*AccountDTO, error) {
	account, err := s.Accounts.GetByID(tx, id)
	if err != nil {
		return nil, err
	}
	return toAccountDTO(account), nil
}

func toAccountDTO(a *models.Account) *AccountDTO {
	return &AccountDTO{
		ID:     a.ID,
		Email:  a.Email.String,
		Status: AccountStatus(a.Status),
	}
}

// verifyArgon2id checks a plaintext password against an argon2id PHC-format hash.
// Expected format: $argon2id$v=19$m=<mem>,t=<time>,p=<par>$<base64-salt>$<base64-hash>
func verifyArgon2id(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, errors.New("invalid argon2id hash format")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("parse version: %w", err)
	}
	if version != argon2.Version {
		return false, fmt.Errorf("unsupported argon2 version: %d", version)
	}

	var memory, iterations, parallelism uint32
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, fmt.Errorf("parse parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}

	computed := argon2.IDKey([]byte(password), salt, iterations, memory, uint8(parallelism), uint32(len(hash)))
	return subtle.ConstantTimeCompare(computed, hash) == 1, nil
}
