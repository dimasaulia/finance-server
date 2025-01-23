package model

import (
	"database/sql"
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
