package model

import "time"

const TB_ROLE string = `"ROLE"`

type Role struct {
	Role_Id   string    `gorm:"primaryKey"`
	Name      string    `gorm:"unique; not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
