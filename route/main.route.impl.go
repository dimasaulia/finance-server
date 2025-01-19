package route

import (
	hc "finance/app/home/controller"
	hr "finance/app/home/router"
	uc "finance/app/user/controller"
	ur "finance/app/user/router"
	us "finance/app/user/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type MainRouter struct {
	App      *fiber.App
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewRoute(app *fiber.App, db *gorm.DB, v *validator.Validate) IMainRouter {
	return &MainRouter{
		App:      app,
		DB:       db,
		Validate: v,
	}
}

func (r *MainRouter) SetupMainRouter() {
	// Home Route
	hr.NewHomeRouter(r.App, hc.NewHomeController()).SetupHomeRouter()

	// User Route
	userService := us.NewUserService(r.DB, r.Validate)
	ur.NewUserRouter(r.App, uc.NewUserController(userService)).SetupUserRouter()
}
