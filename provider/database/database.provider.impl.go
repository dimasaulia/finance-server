package database

import "gorm.io/gorm"

type IPSQLConnetion interface {
	StartPSQLConnection() *gorm.DB
	StartMigration(db *gorm.DB)
}

type PSQLConfiguration struct {
	Host     string
	User     string
	Password string
	Name     string
	Port     string
	SSL      string
}
