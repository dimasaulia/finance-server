package model

import (
	"database/sql"
	"time"
)

const TB_USER string = `"user"`

type User struct {
	IdUser     int64          `gorm:"column:id_user;primaryKey;autoIncrement"`
	Username   string         `gorm:"column:username;unique;not null"`
	Fullname   string         `gorm:"column:fullname"`
	Email      string         `gorm:"column:email;unique;not null"`
	Password   sql.NullString `gorm:"column:password"`
	Provider   string         `gorm:"column:provider"`
	ProviderId sql.NullString `gorm:"column:provider_id;unique"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;autoUpdateTime"`

	// Foriegn Key
	IdRole           sql.NullInt64      `gorm:"column:id_role"`
	Role             Role               `gorm:"foreignKey:id_role;references:id_role"`
	Account          []Account          `gorm:"foreignKey:id_user;references:id_user"`
	TransactionGroup []TransactionGroup `gorm:"foreignKey:id_user;references:id_user"`
	Transaction      []Transaction      `gorm:"foreignKey:id_user;references:id_user"`
	SubTransaction   []SubTransaction   `gorm:"foreignKey:id_user;references:id_user"`
}
