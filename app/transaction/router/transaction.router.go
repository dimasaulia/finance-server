package transaction_router

import (
	tc "finance/app/transaction/controller"
	am "finance/middleware/auth"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
)

type ITransactionRouter interface {
	SetupTransactionRouter()
}

type TransactionRouter struct {
	App        *fiber.App
	Controller tc.ITransactionController
	DB         *gorm.DB
}

func NewTransactionRouter(a *fiber.App, c tc.ITransactionController, db *gorm.DB) ITransactionRouter {
	return &TransactionRouter{
		App:        a,
		Controller: c,
		DB:         db,
	}
}

func (r TransactionRouter) SetupTransactionRouter() {
	log.Info("Setup Transaction Work")
	transactionRouter := r.App.Group("/api/transaction/v1")
	transactionRouter.Use(am.LoginRequired(r.DB))

	transactionRouter.Post("/", r.Controller.CreateNewTransaction)
}
