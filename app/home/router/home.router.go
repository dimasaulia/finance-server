package home

import (
	c "finance/app/home/controller"
	am "finance/middleware/auth"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type IHomeRouter interface {
	SetupHomeRouter()
}

type HomeRouter struct {
	App        *fiber.App
	DB         *gorm.DB
	Controller c.IHomeController
}

func NewHomeRouter(app *fiber.App, c c.IHomeController, db *gorm.DB) IHomeRouter {
	return &HomeRouter{
		App:        app,
		Controller: c,
		DB:         db,
	}
}

func (h *HomeRouter) SetupHomeRouter() {
	h.App.Get("/", h.Controller.Home)
	h.App.Get("/private", am.LoginRequired(h.DB), h.Controller.HomePrivate)
}
