package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

const TB_TRANSACTION_GROUP = "transaction_group"

type TransactionGroup struct {
	IdTransactionGroup int64     `gorm:"column:id_transaction_group;primaryKey;autoIncrement"`
	Description        string    `gorm:"column:transaction"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`
	// Foreign Key
	IdUser int64 `gorm:"column:id_user;foreignKey:id_user;references:id_user"`
	User   User  `gorm:"column:id_user;foreignKey:id_user;references:id_user"`
}

func (t TransactionGroup) AutoCreateTransactionGroup(db *gorm.DB) error {
	var exidtingTransactionGroupCount int64
	db.Model(&t).Select("id_transaction_group").Where("lower(description)", strings.ToLower(t.Description)).Where("id_user", t.IdUser).Count(&exidtingTransactionGroupCount)

	if exidtingTransactionGroupCount == 0 {
		err := db.Create(&t).Error
		if err != nil {
			return err
		}
	}

	return nil
}
