package route

import (
	hc "finance/app/home/controller"
	hr "finance/app/home/router"

	"github.com/gofiber/fiber/v2"
)

type MainRouter struct {
	App *fiber.App
}

func NewRoute(app *fiber.App) IMainRouter {
	return &MainRouter{
		App: app,
	}
}

func (r *MainRouter) SetupMainRouter() {
	hr.NewHomeRouter(r.App, hc.NewHomeController()).SetupHomeRouter()
}
