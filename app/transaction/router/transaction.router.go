package transaction_router

import (
	tc "finance/app/transaction/controller"
	am "finance/middleware/auth"

	"github.com/gofiber/fiber/v2"
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
	transactionRouter := r.App.Group("/api/transaction/v1")
	subTransactionRouter := r.App.Group("/api/sub-transaction/v1")
	transactionRouter.Use(am.LoginRequired(r.DB))
	subTransactionRouter.Use(am.LoginRequired(r.DB))

	transactionRouter.Post("/", r.Controller.CreateNewTransaction)
	transactionRouter.Put("/", r.Controller.UpdateTransaction)
	transactionRouter.Delete("/", r.Controller.DeleteTransaction)
	transactionRouter.Get("/", r.Controller.ListTransaction)

	subTransactionRouter.Post("/", r.Controller.CreateNewSubTransaction)
	subTransactionRouter.Put("/", r.Controller.UpdateSubTransaction)
}
