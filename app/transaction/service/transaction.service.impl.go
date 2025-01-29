package transaction_service

import (
	"database/sql"
	v "finance/app/transaction/validation"
	m "finance/model"
	"finance/utility/generator"
	u "finance/utility/response"
	"fmt"
	"strconv"
	"time"

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

func (t TransactionService) GetUserTransaction(req *u.StandarGetRequest, data *v.UserTransactionDetailRequest) (*[]v.TransactionData, error) {
	var resp []v.TransactionData
	var paramsValue []interface{}
	var paramsKey string = " where "
	currentTime := time.Now()

	err := t.Validator.Struct(data)
	if err != nil {
		return nil, err
	}

	queryTransactionList := `select 
		t.id_transaction,
		t.transaction_code,
		t.transaction_type, 
		t.amount,
		t.balance_before,
		t.balance_after, 
		t.description,
		t.created_at,
		case when t.id_related_transaction is not null then 1 else 0 end "is_have_parent_transaction",
		t.id_related_transaction,
		tg.id_transaction_group,
		tg.description "transaction_name",
		a.id_account,
		a."name" "account_name",
		to_char(t.created_at, 'dd-mm-yyyy')
	from "transaction" t
	inner join transaction_group tg on tg.id_transaction_group = t.id_transaction_group 
	inner join account a on a.id_account = t.id_account`

	paramsKey += " t.id_user = ?"
	paramsValue = append(paramsValue, data.IdUser)

	if *data.IdAccount != "" {
		paramsKey += " and t.id_account = ?"
		paramsValue = append(paramsValue, *data.IdAccount)
	}

	if req.StartDate == "" {
		req.EndDate = currentTime.Format("02-01-2006")
	}

	if req.StartDate == "" {
		req.StartDate = currentTime.AddDate(0, 0, -7).Format("02-01-2006")
	}

	paramsKey += " and to_char(t.created_at, 'dd-mm-yyyy') BETWEEN ? AND ?"
	paramsValue = append(paramsValue, req.StartDate, req.EndDate)

	offset, err := generator.GenerateOffset(req)
	if err != nil {
		return nil, err
	}

	paramsKey += " limit ? offset ?"
	paramsValue = append(paramsValue, strconv.Itoa(offset.Limit), strconv.Itoa(offset.Offset))

	queryTransactionList += paramsKey
	qTransactionList := t.DB.Raw(queryTransactionList, paramsValue...)

	qTransactionList.Scan(&resp)

	if qTransactionList.Error != nil {
		return nil, qTransactionList.Error
	}

	return &resp, nil
}
