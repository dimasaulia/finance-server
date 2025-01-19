package home

import (
	c "finance/app/home/controller"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type IHomeRouter interface {
	SetupHomeRouter()
}

type HomeRouter struct {
	App        *fiber.App
	Controller c.IHomeController
}

func NewHomeRouter(app *fiber.App, c c.IHomeController) IHomeRouter {
	return &HomeRouter{
		App:        app,
		Controller: c,
	}
}

func (h *HomeRouter) SetupHomeRouter() {
	fmt.Println("Setup Home Router")
	h.App.Get("/", h.Controller.Home)
}
