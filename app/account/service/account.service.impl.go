package account_service

import (
	"database/sql"
	"errors"
	av "finance/app/account/validation"
	m "finance/model"
	g "finance/utility/generator"
	r "finance/utility/response"
	"fmt"
	"strings"

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

func (s AccountService) UserAccountList(filter *r.StandarGetRequest, data *av.AccountListRequest) ([]av.AccountCreationResponse, error) {
	var resp []av.AccountCreationResponse
	query := s.DB.Model(&m.Account{}).Select("id_account", "name", "balance", "type")
	query.Where("id_user", data.IdUser)

	if data.Type != "" {
		query.Where("type", data.Type)
	}

	if data.IdAccount != "" {
		query.Where("id_account", data.IdAccount)
	}

	if filter.Search != "" {
		query.Where("lower(name) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Search)))
	}

	offset, err := g.GenerateOffset(filter)
	if err != nil {
		return resp, err
	}

	query.Limit(offset.Limit).Offset(offset.Offset)
	query.Order(fmt.Sprintf("name %v", g.GenerateSort(filter)))
	query.Find(&resp)

	if query.Error != nil {
		return resp, query.Error
	}

	return resp, nil
}

func (s AccountService) DeleteAccountList(id_account string, id_user string) (int64, error) {
	var existingAccountCount int64
	s.DB.Model(&m.Account{}).Where("id_account = ?", id_account).Where("id_user = ?", id_user).Count(&existingAccountCount)

	if existingAccountCount == 0 {
		return 0, errors.New("cant find account")
	}

	query := s.DB.Model(&m.Account{}).Delete("id_account = ?", id_account)
	if query.Error != nil {
		return 0, query.Error
	}

	return query.RowsAffected, nil
}

func (s AccountService) UpdateAccount(data av.AccountUpdateRequest) (av.AccountCreationResponse, error) {
	var resp av.AccountCreationResponse
	err := s.Validator.Struct(data)
	if err != nil {
		return resp, err
	}

	err = m.Account{}.ValidateType(data.Type)
	if err != nil {
		return resp, err
	}

	// Cek Apakah User Memiliki Akun
	var existingAccountCount int64
	err = s.DB.Model(&m.Account{}).Select("id_account").Where("id_account", data.IdAccount).Where("id_user", data.IdUser).Count(&existingAccountCount).Error
	if err != nil {
		return resp, err
	}

	if existingAccountCount == 0 {
		return resp, errors.New("cant find yor account")
	}

	// Update Akun
	updateQuery := s.DB.Save(&m.Account{IdAccount: data.IdAccount, Balance: data.Balance, Name: data.Name, Type: m.Type(data.Type), IdUser: sql.NullInt64{Int64: data.IdUser, Valid: true}}).Where("id_account", data.IdAccount)
	if updateQuery.Error != nil {
		return resp, err
	}

	resp.Balance = data.Balance
	resp.IdAccount = data.IdAccount
	resp.Name = data.Name
	resp.Type = data.Type

	return resp, nil
}
