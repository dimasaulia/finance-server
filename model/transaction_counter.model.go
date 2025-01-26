package model

import "time"

const TB_TRANSACTION_COUNTER = "transaction_counter"

type TransactionCounter struct {
	IdTransactionCounter int64     `gorm:"column:id_transaction_counter;primaryKey;autoIncrement"`
	Counter              int64     `gorm:"column:counter;default=0"`
	Descirption          string    `gorm:"column:descirption"`
	CreatedAt            time.Time `gorm:"autoCreateTime"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime"`

	// Foreign Key
	IdTransactionGroup int64            `gorm:"column:id_transaction_group;foreignKey:id_transaction_group;references:id_transaction_group"`
	TransactionGroup   TransactionGroup `gorm:"foreignKey:id_transaction_group;references:id_transaction_group"`
}
