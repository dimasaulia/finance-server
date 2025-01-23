package account_service

import (
	"errors"
	av "finance/app/account/validation"
	m "finance/model"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	var newAccount m.Account
	// Validasi Request
	err := s.Validator.Struct(req)
	if err != nil {
		return resp, err
	}

	err = m.Account{}.ValidateType(req.Type)
	if err != nil {
		return resp, err
	}

	// Cek apakah user memiliki account dengan nama yang sama sebelumnya
	var existingAccountCount int64
	err = s.DB.Model(&m.Account{}).Where("name = ? AND id_user = ?", req.Name, req.IdUser).Count(&existingAccountCount).Error
	if err != nil {
		return resp, err
	}

	if existingAccountCount > 0 {
		return resp, errors.New("you are already have account with this name")
	}

	// Insert Ke DB
	newAccount.Balance = req.Balance
	newAccount.Name = req.Name
	newAccount.IdUser.Int64 = req.IdUser
	newAccount.IdUser.Valid = true
	newAccount.Type = m.Type(req.Type)

	res := s.DB.Clauses(clause.Returning{}).Select("*").Create(&newAccount)
	if res.Error != nil {
		return resp, res.Error
	}

	resp.Balance = req.Balance
	resp.IdAccount = newAccount.IdAccount
	resp.Name = req.Name
	resp.Type = req.Type

	return resp, nil
}
