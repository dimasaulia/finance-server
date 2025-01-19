package model

import "time"

const TB_ROLE string = `"ROLE"`

type Role struct {
	Id_Role   int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"unique; not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	User      []User    `gorm:"foreignKey:id_role;references:id_role"`
}
