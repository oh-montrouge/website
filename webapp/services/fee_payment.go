package services

import (
	"errors"
	"time"

	"github.com/gobuffalo/pop/v6"
	"ohmontrouge/webapp/models"
)

// ErrDuplicatePayment is returned when a fee payment already exists for the
// (account, season) pair.
var ErrDuplicatePayment = errors.New("une cotisation existe déjà pour ce musicien et cette saison")

// ErrFeePaymentNotFound is returned when a fee payment ID does not exist.
var ErrFeePaymentNotFound = errors.New("cotisation introuvable")

// FeePaymentDTO is the Membership-context view of a fee payment.
type FeePaymentDTO struct {
	ID          int64
	AccountID   int64
	SeasonID    int64
	SeasonLabel string
	Amount      float64
	PaymentDate time.Time
	PaymentType string
	Comment     string
}

// FeePaymentManager is the interface FeePaymentsHandler and MusiciansHandler depend on.
type FeePaymentManager interface {
	Record(tx *pop.Connection, accountID, seasonID int64, amount float64, paymentDate time.Time, paymentType, comment string) error
	Update(tx *pop.Connection, id int64, amount float64, paymentDate time.Time, paymentType, comment string) error
	Delete(tx *pop.Connection, id int64) error
	ListByAccount(tx *pop.Connection, accountID int64) ([]FeePaymentDTO, error)
	GetByID(tx *pop.Connection, id int64) (*FeePaymentDTO, error)
	GetFirstInscriptionDate(tx *pop.Connection, accountID int64) (*time.Time, error)
}

// FeePaymentService implements domain logic for fee payment operations.
type FeePaymentService struct {
	FeePayments FeePaymentRepository
}

func (s FeePaymentService) Record(tx *pop.Connection, accountID, seasonID int64, amount float64, paymentDate time.Time, paymentType, comment string) error {
	_, err := s.FeePayments.Create(tx, accountID, seasonID, amount, paymentDate, paymentType, comment)
	if err != nil {
		if errors.Is(err, models.ErrFeePaymentDuplicate) {
			return ErrDuplicatePayment
		}
		return err
	}
	return nil
}

func (s FeePaymentService) Update(tx *pop.Connection, id int64, amount float64, paymentDate time.Time, paymentType, comment string) error {
	return s.FeePayments.Update(tx, id, amount, paymentDate, paymentType, comment)
}

func (s FeePaymentService) Delete(tx *pop.Connection, id int64) error {
	return s.FeePayments.Delete(tx, id)
}

func (s FeePaymentService) ListByAccount(tx *pop.Connection, accountID int64) ([]FeePaymentDTO, error) {
	rows, err := s.FeePayments.ListByAccount(tx, accountID)
	if err != nil {
		return nil, err
	}
	dtos := make([]FeePaymentDTO, len(rows))
	for i, r := range rows {
		dtos[i] = rowToDTO(r)
	}
	return dtos, nil
}

func (s FeePaymentService) GetByID(tx *pop.Connection, id int64) (*FeePaymentDTO, error) {
	row, err := s.FeePayments.GetByID(tx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrFeePaymentNotFound
	}
	dto := rowToDTO(*row)
	return &dto, nil
}

func (s FeePaymentService) GetFirstInscriptionDate(tx *pop.Connection, accountID int64) (*time.Time, error) {
	return s.FeePayments.GetFirstInscriptionDate(tx, accountID)
}

func rowToDTO(r models.FeePaymentRow) FeePaymentDTO {
	comment := ""
	if r.Comment.Valid {
		comment = r.Comment.String
	}
	return FeePaymentDTO{
		ID:          r.ID,
		AccountID:   r.AccountID,
		SeasonID:    r.SeasonID,
		SeasonLabel: r.SeasonLabel,
		Amount:      r.Amount,
		PaymentDate: r.PaymentDate,
		PaymentType: r.PaymentType,
		Comment:     comment,
	}
}
