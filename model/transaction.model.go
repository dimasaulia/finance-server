package model

import (
	"errors"
	"strings"
	"time"
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
