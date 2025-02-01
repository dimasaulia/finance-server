package model

import (
	"database/sql"
	"errors"
	v "finance/app/transaction/validation"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const TB_SUB_TRANSACTION string = "sub_transaction"

type SubTransaction struct {
	IdSubTransaction int64           `gorm:"column:id_sub_transaction;primaryKey;autoIncement"`
	TransactionType  TransactionType `gorm:"column:transaction_type"`
	TransactionCode  string          `gorm:"column:transaction_code"`
	Amount           float64         `gorm:"column:amount;type:decimal(15,2)"`
	BalanceBefore    float64         `gorm:"column:balance_before;type:decimal(15,2)"`
	BalanceAfter     float64         `gorm:"column:balance_after;type:decimal(15,2)"`
	CreatedAt        time.Time       `gorm:"autoCreateTime"`
	UpdatedAt        time.Time       `gorm:"autoUpdateTime"`
	Description      sql.NullString  `gorm:"column:description"`

	// Foreign Key
	IdTransactionGroup      int64            `gorm:"column:id_transaction_group;foreignKey:id_transaction_group;references:id_transaction_group"`
	TransactionGroup        TransactionGroup `gorm:"foreignKey:id_transaction_group;references:id_transaction_group"`
	IdUser                  int64            `gorm:"column:id_user;foreignKey:id_user;references:id_user"`
	User                    User             `gorm:"column:id_user;foreignKey:id_user;references:id_user"`
	IdTransaction           int64            `gorm:"column:id_transaction;foreignKey:id_transaction;references:id_transaction"`
	Transaction             Transaction      `gorm:"column:id_transaction;foreignKey:id_transaction;references:id_transaction"`
	IdRelatedSubTransaction sql.NullInt64    `gorm:"column:id_related_sub_transaction;foreignKey:id_related_sub_transaction;references:id_sub_transaction"`
	RelatedSubTransaction   []SubTransaction `gorm:"foreignKey:id_related_sub_transaction;references:id_sub_transaction"`
}

func (t SubTransaction) ValidateTransactionType() error {
	switch TransactionType(strings.ToUpper(string(t.TransactionType))) {
	case Debit, Credit:
		return nil
	default:
		return errors.New("transaction type not allow")
	}
}

func (t *SubTransaction) NewTransactionResponse() *v.TransactionResponse {
	return &v.TransactionResponse{
		TransactionCode: t.TransactionCode,
		Amount:          t.Amount,
		BalanceBefore:   t.BalanceBefore,
		BalanceAfter:    t.BalanceAfter,
	}
}

func (t *SubTransaction) CreateNewSubTransaction(tx *gorm.DB) error {
	// Cek Parent Transaction
	parentTransaction := new(Transaction)
	qParentTransaction := tx.Model(&Transaction{}).Where("id_transaction", t.IdTransaction).Where("id_user", t.IdUser).First(&parentTransaction)
	if qParentTransaction.Error != nil {
		return fmt.Errorf("failed to find your transaction. %s", qParentTransaction.Error.Error())
	}

	if parentTransaction.TransactionType == Credit {
		return fmt.Errorf("sorry, this operation is not available. sub-transactions or detail transactions are not supported for credit transactions")
	}

	// Ambil Semua Sub Transaction
	subTransaction := new([]SubTransaction)
	qSubTransaction := tx.Model(&SubTransaction{}).Where("id_transaction", parentTransaction.IdTransaction).Where("id_user", parentTransaction.IdUser).Scan(subTransaction)
	if qSubTransaction.Error != nil {
		return fmt.Errorf("failed to find your transaction detail. %s", qSubTransaction.Error.Error())
	}

	// Create Transaction Code
	err := t.TransactionGroup.AutoCreateTransactionGroup(tx)
	if err != nil {
		return err
	}

	var tgCounter TransactionCounter
	err = tx.Model(&TransactionCounter{}).Select("*").Where("id_transaction_group", t.TransactionGroup.IdTransactionGroup).First(&tgCounter).Error
	if err != nil {
		return fmt.Errorf("failed when create your transaction code, %v", err.Error())
	}

	var now = time.Now()
	var loopCounterFinder bool = true
	var transactionInitial string
	if t.TransactionType == Debit {
		transactionInitial = "D"
	} else {
		transactionInitial = "C"
	}

	var transactionCode string = fmt.Sprintf("%s%s%03d%02d%02d%v", transactionInitial, tgCounter.Descirption, tgCounter.Counter, now.Day(), now.Month(), now.Year())
	for loopCounterFinder {
		var existingTransaction int64
		err := tx.Model(&SubTransaction{}).Select("*").Where("transaction_code", transactionCode).Where("id_user", parentTransaction.IdUser).Count(&existingTransaction).Error

		if err != nil {
			loopCounterFinder = false
			return err
		}

		if existingTransaction == 0 {
			loopCounterFinder = false
		} else {
			tgCounter.Counter += 1
			transactionCode = fmt.Sprintf("%s%s%03d%02d%02d%v", transactionInitial, tgCounter.Descirption, tgCounter.Counter, now.Day(), now.Month(), now.Year())
		}
	}

	// AMBIL AMOUNT BEFORE AND AMOUNT AFTER
	// Ambil amount Parent transaction
	// Lakukan perhitungan untuk mendapatkan amount before and after
	for _, v := range *subTransaction {
		// Jika Tipe Adalah Debit, maka amount parent transaction akan dikurangi
		if v.TransactionType == Debit {
			parentTransaction.Amount = parentTransaction.Amount - v.Amount
		}
		// Jika Tipe Adalah Credit, maka amount parent transaction akan ditambah
		if v.TransactionType == Credit {
			parentTransaction.Amount = parentTransaction.Amount + v.Amount
		}
	}
	t.BalanceBefore = parentTransaction.Amount

	if t.Amount > parentTransaction.Amount {
		return fmt.Errorf("insufficient account balance for the requested debit transaction")
	}

	if t.TransactionType == Debit {
		parentTransaction.Amount = parentTransaction.Amount - t.Amount
	}

	if t.TransactionType == Credit {
		parentTransaction.Amount = parentTransaction.Amount + t.Amount
	}

	log.Infof("Parent ID %v, Amount %v", parentTransaction.IdAccount, parentTransaction.Amount)
	log.Infof("Amount Before: %v", t.BalanceBefore)

	t.BalanceAfter = parentTransaction.Amount
	t.TransactionCode = transactionCode
	t.IdTransactionGroup = t.TransactionGroup.IdTransactionGroup

	// Create sub transaction
	err = tx.Create(&t).Clauses(clause.Returning{}).Error
	if err != nil {
		return fmt.Errorf("failed to create your detail transaction. %s", err.Error())
	}

	// Update Counter
	tgCounter.Counter = tgCounter.Counter + 1
	err = tx.Save(&tgCounter).Where("id_transaction_counter", tgCounter.IdTransactionCounter).Where("id_transaction_group", t.TransactionGroup.IdTransactionGroup).Error
	if err != nil {
		return err
	}
	return nil
}

func (t *SubTransaction) UpdateSubTransaction(tx *gorm.DB) error {
	// Cek Apakah Sub Transaksinya Masih Ada
	subTransaction := new(SubTransaction)
	qSubTransaction := tx.Model(&SubTransaction{}).Select("*").Where("id_sub_transaction", t.IdSubTransaction).Where("id_user", t.IdUser).First(subTransaction)
	if qSubTransaction.Error != nil {
		return fmt.Errorf("we couldn't process your sub transaction. %v", qSubTransaction.Error)
	}
	if qSubTransaction.RowsAffected == 0 {
		return fmt.Errorf("the sub-transaction or detail transaction you requested does not exist or cannot be found")
	}

	// Sekarang Cek Parent Transaction
	parentTransaction := new(Transaction)
	qParentTransaction := tx.Model(&Transaction{}).Select("*").Where("id_transaction", subTransaction.IdTransaction).Where("id_user", subTransaction.IdUser).First(parentTransaction)
	if qParentTransaction.Error != nil {
		return fmt.Errorf("we couldn't process your transaction. %v", qParentTransaction.Error)
	}
	if qParentTransaction.RowsAffected == 0 {
		return fmt.Errorf("the transaction you requested does not exist or cannot be found")
	}

	// Ambil Seluruh data sub transaksi yang terkakit dengan parent transaksi
	allSubTransaction := new([]SubTransaction)
	qAllSubTransaction := tx.Model(&SubTransaction{}).Select("*").Where("id_transaction", parentTransaction.IdTransaction).Where("id_user", t.IdUser).Scan(allSubTransaction)
	if qAllSubTransaction.Error != nil {
		return fmt.Errorf("we couldn't process your transaction. %v", qAllSubTransaction.Error)
	}

	// Lakukan Validasi seluruh data sub transaksi tidak ada sub transaction yang lebih baru dari sub transaksi yang akan diedit
	isNewerSubTransactionExist := false
	for _, subT := range *allSubTransaction {
		if subT.CreatedAt.After(subTransaction.CreatedAt) {
			isNewerSubTransactionExist = true
		}

		// Hanya hitung amount transaksi sebeleum transaksi yang ingin di update
		if subT.CreatedAt.Before(subTransaction.CreatedAt) && subT.IdSubTransaction != t.IdSubTransaction {
			if subT.TransactionType == Credit {
				parentTransaction.Amount += subT.Amount
			}

			if subT.TransactionType == Debit {
				parentTransaction.Amount -= subT.Amount
			}
		}
	}
	if isNewerSubTransactionExist {
		return fmt.Errorf("this sub transaction cannot be modified because a newer transaction exists. please delete the latest sub transaction before making changes")
	}

	if t.Amount > parentTransaction.Amount && t.TransactionType == Debit {
		return fmt.Errorf("insufficient account balance for the requested related credit on related transaction")
	}

	// Hitung amount before dan after berdasarkan sub transaksi sebelumnya
	subTransaction.BalanceBefore = parentTransaction.Amount
	t.BalanceBefore = parentTransaction.Amount
	if t.TransactionType == Credit {
		parentTransaction.Amount += t.Amount
	}
	if t.TransactionType == Debit {
		parentTransaction.Amount -= t.Amount
	}
	subTransaction.BalanceAfter = parentTransaction.Amount
	t.BalanceAfter = parentTransaction.Amount
	subTransaction.Amount = t.Amount
	if t.Description.Valid {
		subTransaction.Description = t.Description
	}
	err := t.TransactionGroup.AutoCreateTransactionGroup(tx)
	if err != nil {
		return err
	}
	subTransaction.TransactionType = t.TransactionType
	subTransaction.IdTransactionGroup = t.IdTransactionGroup

	// Update sub transaction
	err = tx.Save(&subTransaction).Error
	if err != nil {
		return err
	}
	return nil
}
