package route

import (
	account_controller "finance/app/account/controller"
	ar "finance/app/account/router"
	as "finance/app/account/service"
	hc "finance/app/home/controller"
	hr "finance/app/home/router"
	uc "finance/app/user/controller"
	ur "finance/app/user/router"
	us "finance/app/user/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type MainRouter struct {
	App      *fiber.App
	DB       *gorm.DB
	Validate *validator.Validate
	Config   *viper.Viper
}

func NewRoute(app *fiber.App, db *gorm.DB, v *validator.Validate, c *viper.Viper) IMainRouter {
	return &MainRouter{
		App:      app,
		DB:       db,
		Validate: v,
		Config:   c,
	}
}

func (r *MainRouter) SetupMainRouter() {
	// Get ENV DATA
	googleClinetId := r.Config.GetString("GOOGLE_CLIENT_ID")
	googleClinetSecret := r.Config.GetString("GOOGLE_CLIENT_SECRET")
	serverUrl := r.Config.GetString("SERVER_HOST")
	// Home Route
	hr.NewHomeRouter(r.App, hc.NewHomeController(), r.DB).SetupHomeRouter()

	// User Route
	userServiceData := us.UserServiceAdditionalData{
		GoogleClinetId:     googleClinetId,
		GoogleClinetSecret: googleClinetSecret,
		ServerUrl:          serverUrl,
	}
	userService := us.NewUserService(r.DB, r.Validate, userServiceData)
	ur.NewUserRouter(r.App, uc.NewUserController(userService)).SetupUserRouter()

	// Account Router
	accountService := as.NewAccountService(r.DB, r.Validate)
	ar.NewAccountRouter(r.App, account_controller.NewAccountController(accountService), r.DB).SetupAccountRouter()
}
