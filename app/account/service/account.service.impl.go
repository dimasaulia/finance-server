package account_service

import (
	av "finance/app/account/validation"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type AccountService struct {
	DB        *gorm.DB
	Validator *validator.Validate
}

func NewAccountService(db *gorm.DB, v *validator.Validate) IAccountService {
	return &AccountService{
		DB:        db,
		Validator: v,
	}
}

func (s AccountService) CreateAccount(req av.AccountCreationRequest) (av.AccountCreationResponse, error) {
	var resp av.AccountCreationResponse
	return resp, nil
}
