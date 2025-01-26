package model

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
)

type TransactionType string

const (
	TB_TRANSACTION string          = "TRANSACTION"
	Debit          TransactionType = "DEBIT"
	Credit         TransactionType = "CREDIT"
)

type Transaction struct {
	IdTransaction   int64           `gorm:"column:id_transaction;primaryKey;autoIncement"`
	TransactionType TransactionType `gorm:"column:transaction_type"`
	TransactionCode string          `gorm:"column:transaction_code"`
	Amount          float64         `gorm:"column:amount;type:decimal(15,2)"`
	BalanceBefore   float64         `gorm:"column:balance_before;type:decimal(15,2)"`
	BalanceAfter    float64         `gorm:"column:balance_after;type:decimal(15,2)"`
	CreatedAt       time.Time       `gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime"`

	// Foreign Key
	IdTransactionGroup int64            `gorm:"column:id_transaction_group;foreignKey:id_transaction_group;references:id_transaction_group"`
	TransactionGroup   TransactionGroup `gorm:"foreignKey:id_transaction_group;references:id_transaction_group"`
	IdUser             int64            `gorm:"column:id_user;foreignKey:id_user;references:id_user"`
	User               User             `gorm:"column:id_user;foreignKey:id_user;references:id_user"`
	IdAccount          int64            `gorm:"column:id_account;foreignKey:id_account;references:id_account"`
	Account            Account          `gorm:"foreignKey:id_account;references:id_account"`
}

func (t Transaction) ValidateTransactionType() error {
	switch TransactionType(strings.ToUpper(string(t.TransactionType))) {
	case Debit, Credit:
		return nil
	default:
		return errors.New("transaction type not allow")
	}
}

func (t *Transaction) CrateNewTransaction(db *gorm.DB) error {
	// Cek Transaction Group
	if t.TransactionGroup.Description == "" {
		return errors.New("please fill transaction group")
	}

	err := t.TransactionGroup.AutoCreateTransactionGroup(db)
	if err != nil {
		return err
	}

	// Ambil Data Counter
	var tgCounter TransactionCounter
	err = db.Model(&TransactionCounter{}).Select("*").Where("id_transaction_group", t.TransactionGroup.IdTransactionGroup).First(&tgCounter).Error
	if err != nil {
		return fmt.Errorf("failed when get counter, %v", err.Error())
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
	log.Infof("Counter => %v", transactionCode)
	for loopCounterFinder {
		var existingTransaction int64
		err := db.Model(&Transaction{}).Select("*").Where("transaction_code", transactionCode).Count(&existingTransaction).Error

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

	// Cek Account dan ambil amount
	var userAccount Account
	qAccount := db.Model(&Account{}).Select("*").Where("id_account", t.IdAccount).Where("id_user", t.IdUser).First(&userAccount)
	if qAccount.Error != nil {
		return qAccount.Error
	}

	// DEBIT => SALDO BERKURANG; CREDIT => SALDO BERTAMBAH
	if t.TransactionType == Debit && userAccount.Balance < t.Amount {
		return errors.New("insufficient account balance for the requested debit transaction")
	}
	t.BalanceBefore = userAccount.Balance

	// Kurangi atau tambah amount, serta lakukan update amount pada tabel account
	if t.TransactionType == Debit {
		userAccount.Balance = userAccount.Balance - t.Amount
	}

	if t.TransactionType == Credit {
		userAccount.Balance = userAccount.Balance + t.Amount
	}

	t.IdTransactionGroup = t.TransactionGroup.IdTransactionGroup
	t.IdAccount = userAccount.IdAccount
	t.BalanceAfter = userAccount.Balance
	t.TransactionCode = transactionCode

	// Open DB Transaction
	tx := db.Begin()

	// Create transaction
	err = tx.Create(&t).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Update Account, khususnya ammount
	err = tx.Save(&userAccount).Where("id_account", userAccount.IdAccount).Where("id_user", userAccount.IdUser).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Update Counter
	err = tx.Save(&tgCounter).Where("id_transaction_counter", tgCounter.IdTransactionCounter).Where("id_transaction_group", t.TransactionGroup.IdTransactionGroup).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
