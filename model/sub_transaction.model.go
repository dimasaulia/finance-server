package model

import (
	"database/sql"
	"errors"
	"strings"
	"time"
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
	IdTransaction           int64            `gorm:"column:id_transaction;foreignKey:id_user;references:id_transaction"`
	Transaction             Transaction      `gorm:"column:id_transaction;foreignKey:id_user;references:id_transaction"`
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
