package model

import "time"

const TB_ROLE string = `"ROLE"`

type Role struct {
	Id_Role   string    `gorm:"primaryKey"`
	Name      string    `gorm:"unique; not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
