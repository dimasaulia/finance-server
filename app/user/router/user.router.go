package user

import (
	c "finance/app/user/controller"

	"github.com/gofiber/fiber/v2"
)

type IHomeRouter interface {
	SetupUserRouter()
}

type HomeRouter struct {
	App        *fiber.App
	Controller c.IUserController
}

func NewUserRouter(app *fiber.App, c c.IUserController) IHomeRouter {
	return &HomeRouter{
		App:        app,
		Controller: c,
	}
}

func (h *HomeRouter) SetupUserRouter() {
	userV1 := h.App.Group("/api/user/v1")
	userV1.Post("/", h.Controller.ManualRegistration)
	userV1.Post("/login", h.Controller.ManualLogin)
}
