package database

import "gorm.io/gorm"

type IPSQLConnetion interface {
	StartPSQLConnection() *gorm.DB
}

type PSQLConfiguration struct {
	Host     string
	User     string
	Password string
	Name     string
	Port     string
	SSL      string
}
