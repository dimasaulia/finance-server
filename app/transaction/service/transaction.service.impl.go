package transaction_service

import (
	"database/sql"
	v "finance/app/transaction/validation"
	m "finance/model"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type TransactionService struct {
	DB        *gorm.DB
	Validator *validator.Validate
}

func NewTransactionService(db *gorm.DB, v *validator.Validate) ITransactionService {
	return &TransactionService{
		DB:        db,
		Validator: v,
	}
}

func (t TransactionService) CreateNewTransaction(req *v.NewTransactionRequest) (*[]v.TransactionResponse, error) {
	var resp []v.TransactionResponse
	// Validasi Request
	err := t.Validator.Struct(req)
	if err != nil {
		return nil, err
	}

	var sourceAmount float64 = req.Amount
	if req.AdminFee != nil {
		sourceAmount = req.Amount + *req.AdminFee
	}

	sourceTransaction := m.Transaction{
		Amount:          sourceAmount,
		TransactionType: m.TransactionType(req.TransactionType),
		IdUser:          req.IdUser,
		IdAccount:       req.IdAccount,
		TransactionGroup: m.TransactionGroup{
			IdUser:      req.IdUser,
			Description: req.TransactionGroup,
		},
	}

	if req.Description != nil {
		sourceTransaction.Description.String = *req.Description
		sourceTransaction.Description.Valid = true
	}

	err = sourceTransaction.ValidateTransactionType()
	if err != nil {
		return nil, err
	}

	if req.IdAccountDestination != nil && req.IdAccount == *req.IdAccountDestination {
		return nil, fmt.Errorf("source account and destination account cannot be the same")
	}

	tx := t.DB.Begin()
	err = sourceTransaction.CrateNewTransaction(tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Jika transaksi adalah pemindahakn kekayaan antar akun
	// Maka akun tujuan akan membuat transaksi kredit
	if req.IdAccountDestination != nil && req.TransactionType == "DEBIT" && req.IdAccount != *req.IdAccountDestination {
		destinationTransaction := m.Transaction{
			Amount:               req.Amount,
			TransactionType:      m.Credit,
			IdUser:               req.IdUser,
			IdAccount:            *req.IdAccountDestination,
			Description:          sourceTransaction.Description,
			IdRelatedTransaction: sql.NullInt64{Int64: sourceTransaction.IdTransaction, Valid: true},
			TransactionGroup: m.TransactionGroup{
				IdUser:      req.IdUser,
				Description: req.TransactionGroup,
			},
		}

		err = destinationTransaction.CrateNewTransaction(tx)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		resp = append(resp, *destinationTransaction.NewTransactionResponse())
	}

	tx.Commit()
	resp = append(resp, *sourceTransaction.NewTransactionResponse())

	return &resp, nil
}

func (t TransactionService) UpdateTransaction(req *v.UpdateTransactionRequest) (*[]v.TransactionResponse, error) {
	// Validasi Request
	err := t.Validator.Struct(req)
	if err != nil {
		return nil, err
	}

	transaction := m.Transaction{
		IdTransaction:   req.IdTransaction,
		Amount:          req.Amount,
		TransactionType: m.TransactionType(req.TransactionType),
		IdUser:          req.IdUser,
		TransactionGroup: m.TransactionGroup{
			IdUser:      req.IdUser,
			Description: req.TransactionGroup,
		},
	}
	if req.Description != nil {
		transaction.Description.Valid = true
		transaction.Description.String = *req.Description
	}

	err = transaction.ValidateTransactionType()
	if err != nil {
		return nil, err
	}

	tx := t.DB.Begin()
	var adminFee sql.NullFloat64
	if req.AdminFee != nil {
		adminFee = sql.NullFloat64{Valid: true, Float64: *req.AdminFee}
	}
	resp, err := transaction.UpdateExistingTransaction(tx, req.IdAccountDestination, adminFee)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return resp, nil
}

func (t TransactionService) DeleteTransaction(req *v.DeleteTransactionRequest) error {
	// Validasi Request
	err := t.Validator.Struct(req)
	if err != nil {
		return err
	}

	delatedTransaction := m.Transaction{
		IdTransaction: req.IdTransaction,
		IdUser:        req.IdUser,
	}

	tx := t.DB.Begin()
	err = delatedTransaction.DeleteTransaction(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
