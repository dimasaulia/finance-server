package transaction_service

import (
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

	sourceTransaction := m.Transaction{
		Amount:          req.Amount,
		TransactionType: m.TransactionType(req.TransactionType),
		IdUser:          req.IdUser,
		IdAccount:       req.IdAccount,
		TransactionGroup: m.TransactionGroup{
			IdUser:      req.IdUser,
			Description: req.TransactionGroup,
		},
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
			Amount:          req.Amount,
			TransactionType: m.Credit,
			IdUser:          req.IdUser,
			IdAccount:       *req.IdAccountDestination,
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
