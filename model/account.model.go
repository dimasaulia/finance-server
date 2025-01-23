package model

import (
	"database/sql"
	"errors"
	"time"
)

type Type string

const (
	TB_ACCOUNT  string = `"ACCOUNT"`
	Bank        Type   = "BANK"
	EWallet     Type   = "EWALLET"
	Investation Type   = "INVESTATION"
	Other       Type   = "OTHER"
)

type Account struct {
	IdAccount int64     `gorm:"primaryKey;autoIncrement;column:id_account"`
	Name      string    `gorm:"column:name"`
	Balance   float64   `gorm:"column:balance"`
	Type      Type      `gorm:"column:type"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	IdUser sql.NullInt64 `gorm:"foreignKey:id_user;references:id_user"`
	User   User          `gorm:"foreignKey:id_user;references:id_user"`
}

func (a Account) ValidateType(t string) error {
	switch Type(t) {
	case Bank, EWallet, Investation, Other:
		return nil
	default:
		return errors.New("account type not allow")
	}
}
