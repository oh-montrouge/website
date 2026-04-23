package services

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

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
	ErrInvalidToken       = errors.New("token not found, expired, or already used")
)

// InviteContextDTO carries account info needed to render the invite form.
// Returned by ValidateInviteToken; the handler passes it to the template.
type InviteContextDTO struct {
	TokenID        int64
	AccountID      int64
	FirstName      string
	LastName       string
	Email          string
	InstrumentName string
}

// PasswordResetContextDTO carries the token and account IDs needed to complete a reset.
// Returned by ValidatePasswordResetToken.
type PasswordResetContextDTO struct {
	TokenID   int64
	AccountID int64
}

// InviteTokenDTO is returned by GenerateInviteToken for display in the admin UI.
type InviteTokenDTO struct {
	Token     string
	URL       string
	ExpiresAt time.Time
}

// PasswordResetTokenDTO is returned by GeneratePasswordResetToken for display in the admin UI.
type PasswordResetTokenDTO struct {
	Token     string
	URL       string
	ExpiresAt time.Time
}

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

// AccountTokenManager is the interface TokensHandler depends on for invite and reset flows.
type AccountTokenManager interface {
	ValidateInviteToken(tx *pop.Connection, token string) (*InviteContextDTO, error)
	CompleteInvite(tx *pop.Connection, tokenID, accountID int64, passwordHash string, phoneAddressConsent bool) error
	ValidatePasswordResetToken(tx *pop.Connection, token string) (*PasswordResetContextDTO, error)
	CompletePasswordReset(tx *pop.Connection, tokenID, accountID int64, passwordHash string) error
}

// AccountService implements domain logic for account operations.
type AccountService struct {
	Accounts     AccountRepository
	Roles        RoleRepository
	InviteTokens InviteTokenRepository
	ResetTokens  PasswordResetTokenRepository
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

// GenerateInviteToken creates a new invite token for the account, invalidating any existing one.
// baseURL is the application's public root (e.g. "https://ohm.example.com"); it is used to
// build the full invite URL returned in the DTO.
func (s AccountService) GenerateInviteToken(tx *pop.Connection, accountID int64, baseURL string) (*InviteTokenDTO, error) {
	token, err := generateCSPRNGToken()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)
	if err := s.InviteTokens.InvalidateExisting(tx, accountID); err != nil {
		return nil, err
	}
	if err := s.InviteTokens.Generate(tx, accountID, token, expiresAt); err != nil {
		return nil, err
	}
	return &InviteTokenDTO{
		Token:     token,
		URL:       baseURL + "/invitation/" + token,
		ExpiresAt: expiresAt,
	}, nil
}

// ValidateInviteToken validates the token and returns account info for rendering the invite form.
// Returns ErrInvalidToken when the token is not found, expired, used, or the account is not pending.
func (s AccountService) ValidateInviteToken(tx *pop.Connection, token string) (*InviteContextDTO, error) {
	row, err := s.InviteTokens.FindByToken(tx, token)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrInvalidToken
	}
	return &InviteContextDTO{
		TokenID:        row.TokenID,
		AccountID:      row.AccountID,
		FirstName:      row.FirstName,
		LastName:       row.LastName,
		Email:          row.Email,
		InstrumentName: row.InstrumentName,
	}, nil
}

// CompleteInvite activates the account and marks the invite token used, atomically.
// Both operations run on the same tx; no RSVP seeding — that is wired in Phase 4.6.
func (s AccountService) CompleteInvite(tx *pop.Connection, tokenID, accountID int64, passwordHash string, phoneAddressConsent bool) error {
	if err := s.Accounts.Activate(tx, accountID, passwordHash, phoneAddressConsent); err != nil {
		return err
	}
	return s.InviteTokens.MarkUsed(tx, tokenID)
}

// GeneratePasswordResetToken creates a new reset token for an active account, invalidating any existing one.
func (s AccountService) GeneratePasswordResetToken(tx *pop.Connection, accountID int64, baseURL string) (*PasswordResetTokenDTO, error) {
	token, err := generateCSPRNGToken()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)
	if err := s.ResetTokens.InvalidateExisting(tx, accountID); err != nil {
		return nil, err
	}
	if err := s.ResetTokens.Generate(tx, accountID, token, expiresAt); err != nil {
		return nil, err
	}
	return &PasswordResetTokenDTO{
		Token:     token,
		URL:       baseURL + "/reinitialiser-mot-de-passe/" + token,
		ExpiresAt: expiresAt,
	}, nil
}

// ValidatePasswordResetToken validates the token and returns IDs needed to complete the reset.
// Returns ErrInvalidToken when the token is not found, expired, used, or the account is not active.
func (s AccountService) ValidatePasswordResetToken(tx *pop.Connection, token string) (*PasswordResetContextDTO, error) {
	row, err := s.ResetTokens.FindByToken(tx, token)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrInvalidToken
	}
	return &PasswordResetContextDTO{
		TokenID:   row.TokenID,
		AccountID: row.AccountID,
	}, nil
}

// CompletePasswordReset updates the account's password and marks the token used, atomically.
func (s AccountService) CompletePasswordReset(tx *pop.Connection, tokenID, accountID int64, passwordHash string) error {
	if err := s.Accounts.UpdatePasswordHash(tx, accountID, passwordHash); err != nil {
		return err
	}
	return s.ResetTokens.MarkUsed(tx, tokenID)
}

// ValidatePasswordStrength enforces the password policy and confirms the two fields match.
// Policy: minimum 22 characters, at least one uppercase, one lowercase, one digit, one special character.
func ValidatePasswordStrength(password, confirm string) error {
	if password != confirm {
		return errors.New("les mots de passe ne correspondent pas")
	}
	if len(password) < 22 {
		return errors.New("le mot de passe doit contenir au moins 22 caractères")
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return errors.New("le mot de passe doit contenir au moins une majuscule, une minuscule, un chiffre et un caractère spécial")
	}
	return nil
}

func generateCSPRNGToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
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
