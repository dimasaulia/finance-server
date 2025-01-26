package model

import (
	"errors"
	"strings"
	"time"

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

	err = db.Save(&userAccount).Where("id_account", userAccount.IdAccount).Where("id_user", userAccount.IdUser).Error
	if err != nil {
		return err
	}

	t.IdTransactionGroup = t.TransactionGroup.IdTransactionGroup
	t.IdAccount = userAccount.IdAccount
	t.BalanceAfter = userAccount.Balance

	// Create transaction
	err = db.Create(&t).Error
	if err != nil {
		return err
	}

	return nil
}
