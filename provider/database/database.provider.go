package database

import (
	"finance/model"
	"fmt"

	"github.com/gofiber/fiber/v2/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type PSQL struct {
	Configuration PSQLConfiguration
}

func NewPSQLConnetion(conf PSQLConfiguration) IPSQLConnetion {
	return &PSQL{
		Configuration: conf,
	}
}

func (p *PSQL) StartPSQLConnection() *gorm.DB {
	DSN := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=%v", p.Configuration.Host, p.Configuration.User, p.Configuration.Password, p.Configuration.Name, p.Configuration.Port, p.Configuration.SSL)
	fmt.Printf("START NEW POSGTRE SQL CONNECTION At %v@%v For \"%v\" \n", p.Configuration.User, p.Configuration.Host, p.Configuration.Name)

	PSQLConnection := postgres.Open(DSN)

	gormConfig := gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	}

	psql, err := gorm.Open(PSQLConnection, &gormConfig)

	if err != nil {
		log.Warn("Failed to connect to database")
		panic("Failed to connect to database")
	}

	return psql
}

func (p *PSQL) StartMigration(db *gorm.DB) {
	log.Info("==================== Start Migration All Table ====================")

	db.AutoMigrate(
		&model.Role{},
		&model.User{},
		&model.Account{},
		&model.TransactionGroup{},
		&model.Transaction{},
		&model.TransactionCounter{},
	)

	db.Migrator().CreateConstraint(&model.Role{}, "user")
	db.Migrator().CreateConstraint(&model.Role{}, "fk_role_user")

	db.Migrator().CreateConstraint(&model.User{}, "Account")
	db.Migrator().CreateConstraint(&model.User{}, "fk_user_account")

	db.Migrator().CreateConstraint(&model.User{}, "TransactionGroup")
	db.Migrator().CreateConstraint(&model.User{}, "fk_user_transaction_group")

	// Foreign Key Untuk Tabel Transaction
	db.Migrator().CreateConstraint(&model.User{}, "Transaction")
	db.Migrator().CreateConstraint(&model.User{}, "fk_user_transaction")
	db.Migrator().CreateConstraint(&model.Account{}, "Account")
	db.Migrator().CreateConstraint(&model.Account{}, "fk_account_transaction")
	db.Migrator().CreateConstraint(&model.TransactionGroup{}, "TransactionGroup")
	db.Migrator().CreateConstraint(&model.TransactionGroup{}, "fk_transaction_group_transaction")

	db.Migrator().CreateConstraint(&model.TransactionGroup{}, "TransactionCounter")
	db.Migrator().CreateConstraint(&model.TransactionGroup{}, "fk_transaction_group_transaction_counter")
}
