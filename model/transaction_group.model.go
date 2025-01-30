package model

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const TB_TRANSACTION_GROUP = "transaction_group"

type TransactionGroup struct {
	IdTransactionGroup int64     `gorm:"column:id_transaction_group;primaryKey;autoIncrement"`
	Description        string    `gorm:"column:description"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`
	// Foreign Key
	IdUser             int64                `gorm:"column:id_user;foreignKey:id_user;references:id_user"`
	User               User                 `gorm:"column:id_user;foreignKey:id_user;references:id_user"`
	Transaction        []Transaction        `gorm:"foreignKey:id_transaction_group;references:id_transaction_group"`
	TransactionCounter []TransactionCounter `gorm:"foreignKey:id_transaction_group;references:id_transaction_group"`
	SubTransaction     []SubTransaction     `gorm:"foreignKey:id_transaction_group;references:id_transaction_group"`
}

func (t *TransactionGroup) AutoCreateTransactionGroup(db *gorm.DB) error {
	if t.Description == "" {
		return fmt.Errorf("please fill transaction group")
	}

	t.Description = strings.ToUpper(t.Description)
	tgQuery := db.Model(&t).Select("*").Where("UPPER(description) = ?", (t.Description)).Where("id_user", t.IdUser).First(t)
	if tgQuery.Error != nil && tgQuery.RowsAffected != 0 {
		return tgQuery.Error
	}

	if tgQuery.RowsAffected == 0 {
		err := db.Create(&t).Clauses(clause.Returning{}).Error
		if err != nil {
			return err
		}
	}

	var existingCounter int64
	var newCounter TransactionCounter
	newCounter.Counter = 1
	newCounter.IdTransactionGroup = t.IdTransactionGroup

	descArr := strings.Split(t.Description, " ")
	for _, v := range descArr {
		newCounter.Descirption = fmt.Sprintf("%v%v", newCounter.Descirption, v[:1])
	}

	err := db.Model(&TransactionCounter{}).Select("*").Where("id_transaction_group", t.IdTransactionGroup).Where("descirption", newCounter.Descirption).Count(&existingCounter).Error
	if err != nil {
		return err
	}

	if existingCounter == 0 {
		err := db.Create(&newCounter).Error
		if err != nil {
			return fmt.Errorf("failed to create transaction counter, %v", err)
		}
	}

	return nil
}
