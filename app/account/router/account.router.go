package account_router

import (
	ac "finance/app/account/controller"
	am "finance/middleware/auth"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type IAccountRouter interface {
	SetupAccountRouter()
}

type AccountRouter struct {
	App        *fiber.App
	Controller ac.IAccountController
	DB         *gorm.DB
}

func NewAccountRouter(app *fiber.App, c ac.IAccountController, db *gorm.DB) IAccountRouter {
	return &AccountRouter{
		App:        app,
		Controller: c,
		DB:         db,
	}
}

func (h *AccountRouter) SetupAccountRouter() {
	accountV1 := h.App.Group("/api/account/v1")

	accountV1.Use(am.LoginRequired(h.DB))
	accountV1.Post("/", h.Controller.CreateNewAccount)
	accountV1.Put("/", h.Controller.UpdateAccount)
	accountV1.Get("/", h.Controller.UserAccount)
	accountV1.Delete("/:id", h.Controller.DeleteAccount)
}
