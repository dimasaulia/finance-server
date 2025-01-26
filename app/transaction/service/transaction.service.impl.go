package transaction_service

import (
	v "finance/app/transaction/validation"
	m "finance/model"

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

func (t TransactionService) CreateNewTransaction(req *v.NewTransactionRequest) (*v.TransactionResponse, error) {
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

	err = sourceTransaction.CrateNewTransaction(t.DB)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
