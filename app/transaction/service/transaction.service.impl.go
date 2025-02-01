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

func (t *TransactionService) CreateNewTransaction(req *v.NewTransactionRequest) (*[]v.TransactionResponse, error) {
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

func (t *TransactionService) UpdateTransaction(req *v.UpdateTransactionRequest) (*[]v.TransactionResponse, error) {
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

func (t *TransactionService) DeleteTransaction(req *v.DeleteTransactionRequest) error {
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

func (t *TransactionService) GetUserTransaction(req *u.StandarGetRequest, data *v.UserTransactionDetailRequest) (*[]v.TransactionData, error) {
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
        t.id_related_transaction "id_parent_transaction",
		case when rt.id_related_transaction is not null then 1 else 0 end "is_have_child_transaction",
		rt.id_transaction "id_child_transaction",
		tg.id_transaction_group,
		tg.description "transaction_name",
		a.id_account,
		a."name" "account_name",
		to_char(t.created_at, 'dd-mm-yyyy')
	from "transaction" t
	inner join transaction_group tg on tg.id_transaction_group = t.id_transaction_group 
	inner join account a on a.id_account = t.id_account
	left  join "transaction" rt on rt.id_related_transaction = t.id_transaction`

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

	paramsKey += " and to_date(to_char(t.created_at, 'dd-mm-yyyy'),'dd-mm-yyyy') BETWEEN to_date(?,'dd-mm-yyyy') AND to_date(?,'dd-mm-yyyy')"
	paramsValue = append(paramsValue, req.StartDate, req.EndDate)

	offset, err := generator.GenerateOffset(req)
	if err != nil {
		return nil, err
	}

	paramsKey += " order by t.id_transaction limit ? offset ?"
	paramsValue = append(paramsValue, strconv.Itoa(offset.Limit), strconv.Itoa(offset.Offset))

	queryTransactionList += paramsKey
	qTransactionList := t.DB.Raw(queryTransactionList, paramsValue...)

	qTransactionList.Scan(&resp)

	if qTransactionList.Error != nil {
		return nil, qTransactionList.Error
	}

	return &resp, nil
}

func (t *TransactionService) CreateNewSubTransaction(req *v.NewSubTransactionRequest) (*[]v.TransactionResponse, error) {
	// Cek dan Validasi Request
	err := t.Validator.Struct(req)
	if err != nil {
		return nil, fmt.Errorf("request validation failed: %v. Please check the fields and ensure they match the required format", err.Error())
	}

	if req.IdTransactionDestination != nil && req.IdTransaction == *req.IdTransactionDestination {
		return nil, fmt.Errorf("source transaction and destination transaction cannot be the same")
	}

	resp := new([]v.TransactionResponse)

	// Tambahkan Admin Fee Jika Ada
	var sourceAmount float64 = req.Amount
	if req.AdminFee != nil {
		sourceAmount = sourceAmount + *req.AdminFee
	}

	// Buat Objek
	newSourceSubTransaction := new(m.SubTransaction)
	newSourceSubTransaction.Amount = sourceAmount
	newSourceSubTransaction.TransactionType = m.TransactionType(req.TransactionType)
	newSourceSubTransaction.IdTransaction = req.IdTransaction
	newSourceSubTransaction.IdUser = req.IdUser
	newSourceSubTransaction.TransactionGroup.IdUser = req.IdUser
	newSourceSubTransaction.TransactionGroup.Description = req.TransactionGroup
	if req.Description != nil {
		newSourceSubTransaction.Description.String = *req.Description
		newSourceSubTransaction.Description.Valid = true
	}

	err = newSourceSubTransaction.ValidateTransactionType()
	if err != nil {
		return nil, err
	}

	// Buka transaksi
	tx := t.DB.Begin()
	err = newSourceSubTransaction.CreateNewSubTransaction(tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if req.TransactionType == string(m.Debit) && req.IdTransactionDestination != nil {
		newDestinationSubTransaction := new(m.SubTransaction)
		newDestinationSubTransaction.Amount = req.Amount
		newDestinationSubTransaction.TransactionType = m.Credit
		newDestinationSubTransaction.IdUser = req.IdUser
		newDestinationSubTransaction.IdTransaction = *req.IdTransactionDestination
		newDestinationSubTransaction.TransactionGroup.IdUser = req.IdUser
		newDestinationSubTransaction.TransactionGroup.Description = req.TransactionGroup
		newDestinationSubTransaction.IdRelatedSubTransaction.Int64 = newSourceSubTransaction.IdSubTransaction
		newDestinationSubTransaction.IdRelatedSubTransaction.Valid = true
		if req.Description != nil {
			newDestinationSubTransaction.Description.String = *req.Description
			newDestinationSubTransaction.Description.Valid = true
		}
		err = newDestinationSubTransaction.CreateNewSubTransaction(tx)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	*resp = append(*resp, *newSourceSubTransaction.NewTransactionResponse())
	tx.Commit()
	// Commit transaksi
	return resp, nil
}

func (t *TransactionService) UpdateSubTransaction(req *v.UpdateSubTransactionRequest) (*[]v.TransactionResponse, error) {
	resp := new([]v.TransactionResponse)
	// Validasi Request
	err := t.Validator.Struct(req)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while validating your request. Please review the details and try again. Details: %s", err)
	}

	// Cek Apakah memiliki parent transaction
	subTransaction := new(m.SubTransaction)
	qSubTransaction := t.DB.Model(&m.SubTransaction{}).Select("*").Where("id_sub_transaction", req.IdSubTransaction).Where("id_user", req.IdUser).First(subTransaction)
	if qSubTransaction.Error != nil {
		return nil, fmt.Errorf("we couldn't process your sub transaction. %v", qSubTransaction.Error)
	}
	if subTransaction.IdRelatedSubTransaction.Valid {
		return nil, fmt.Errorf("cannot modify this transaction because it has a parent transaction. To delete this transaction, you must first modify the parent transaction")
	}

	// Cek Apakah Memiliki Sub Transaction
	subRelatedTransaction := new([]m.SubTransaction)
	qSubRelatedTransaction := t.DB.Model(&m.SubTransaction{}).Select("*").Where("id_related_sub_transaction", subTransaction.IdSubTransaction).Where("id_user", req.IdUser).Scan(subRelatedTransaction)
	if qSubRelatedTransaction.Error != nil {
		return nil, fmt.Errorf("we couldn't process your sub transaction. %v", qSubRelatedTransaction.Error)
	}

	// Jika Memiliki sub transaction, tidak bisa melakukan proses credit
	if qSubRelatedTransaction.RowsAffected > 0 && req.TransactionType == string(m.Credit) {
		return nil, fmt.Errorf("transaction type cannot be changed to credit because this transaction has linked sub-transactions")
	}
	// Jika Memiliki sub transaction, id destination transaction harus terisi
	if qSubRelatedTransaction.RowsAffected > 0 && req.IdTransactionDestination == nil {
		return nil, fmt.Errorf("the transaction you are trying to update is linked to other transactions. you must specify a new destination account for this transaction")
	}
	// Jika Memiliki sub transaction, id destination dan id transaction tidak boleh sama
	if qSubRelatedTransaction.RowsAffected > 0 && *req.IdTransactionDestination == subTransaction.IdTransaction {
		return nil, fmt.Errorf("we're unable to process your request. the destination transaction cannot be the same as the current transaction")
	}

	sourceAmount := req.Amount
	if req.AdminFee != nil {
		sourceAmount += *req.AdminFee
	}

	// Buka DB TRX
	tx := t.DB.Begin()
	// Buat Objek Untuk Memodifikasi Sub Transaction
	subTransaction.Amount = sourceAmount
	subTransaction.TransactionGroup.Description = req.TransactionGroup
	subTransaction.TransactionGroup.IdUser = req.IdUser
	if req.Description != nil {
		subTransaction.Description.String = *req.Description
		subTransaction.Description.Valid = true
	}
	subTransaction.TransactionType = m.TransactionType(req.TransactionType)
	err = subTransaction.ValidateTransactionType()
	if err != nil {
		return nil, err
	}
	subTransaction.IdUser = req.IdUser
	err = subTransaction.UpdateSubTransaction(tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	*resp = append(*resp, *subTransaction.NewTransactionResponse())

	// Cari sub transaksi yang memiliki relasi ke transaksi yang ingin di ubah
	for _, subRelated := range *subRelatedTransaction {
		// Jika Destinasi Sama maka update saja menyesuaikan amount baru
		if subRelated.IdTransaction == *req.IdTransactionDestination {
			subRelated.Amount = req.Amount
			subRelated.TransactionGroup.Description = req.TransactionGroup
			subRelated.TransactionGroup.IdUser = req.IdUser
			subRelated.TransactionType = m.Credit
			if req.Description != nil {
				subRelated.Description.String = *req.Description
				subRelated.Description.Valid = true
			}

			err = subRelated.UpdateSubTransaction(tx)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
			*resp = append(*resp, *subRelated.NewTransactionResponse())

		}
		// Jika Destinasi berbeda maka delete transaksi lama dan buat yang baru
		if subRelated.IdTransaction != *req.IdTransactionDestination {
			err = subRelated.DeleteSubTransaction(tx)
			if err != nil {
				tx.Rollback()
				return nil, err
			}

			newDestinationSubTransaction := new(m.SubTransaction)
			newDestinationSubTransaction.Amount = req.Amount
			newDestinationSubTransaction.TransactionType = m.Credit
			newDestinationSubTransaction.IdUser = req.IdUser
			newDestinationSubTransaction.IdTransaction = *req.IdTransactionDestination
			newDestinationSubTransaction.TransactionGroup.IdUser = req.IdUser
			newDestinationSubTransaction.TransactionGroup.Description = req.TransactionGroup
			newDestinationSubTransaction.IdRelatedSubTransaction.Int64 = subTransaction.IdSubTransaction
			newDestinationSubTransaction.IdRelatedSubTransaction.Valid = true
			if req.Description != nil {
				newDestinationSubTransaction.Description.String = *req.Description
				newDestinationSubTransaction.Description.Valid = true
			}
			err = newDestinationSubTransaction.CreateNewSubTransaction(tx)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}
	// Commit DB TRX
	tx.Commit()

	return resp, nil
}

func (t *TransactionService) DeleteSubTransaction(req *v.DeleteSubTransactionRequest) error {
	// Validasi Request
	err := t.Validator.Struct(req)
	if err != nil {
		return err
	}

	delatedTransaction := m.SubTransaction{
		IdSubTransaction: req.IdSubTransaction,
		IdUser:           req.IdUser,
	}

	// Cek Apakah memiliki parent transaction
	subTransaction := new(m.SubTransaction)
	qSubTransaction := t.DB.Model(&m.SubTransaction{}).Select("*").Where("id_sub_transaction", req.IdSubTransaction).Where("id_user", req.IdUser).First(subTransaction)
	if qSubTransaction.Error != nil {
		return fmt.Errorf("we couldn't process your sub transaction. %v", qSubTransaction.Error)
	}
	if subTransaction.IdRelatedSubTransaction.Valid {
		return fmt.Errorf("cannot modify this transaction because it has a parent transaction. To delete this transaction, you must first modify the parent transaction")
	}

	// Cek Apakah Memiliki Sub Transaction
	subRelatedTransaction := new([]m.SubTransaction)
	qSubRelatedTransaction := t.DB.Model(&m.SubTransaction{}).Select("*").Where("id_related_sub_transaction", req.IdSubTransaction).Where("id_user", req.IdUser).Scan(subRelatedTransaction)
	if qSubRelatedTransaction.Error != nil {
		return fmt.Errorf("we couldn't process your sub transaction. %v", qSubRelatedTransaction.Error)
	}

	tx := t.DB.Begin()
	for _, relatedSubTransaction := range *subRelatedTransaction {
		err = relatedSubTransaction.DeleteSubTransaction(tx)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = delatedTransaction.DeleteSubTransaction(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
