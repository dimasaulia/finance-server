package account_controller

import (
	as "finance/app/account/service"

	"github.com/gofiber/fiber/v2"
)

type AccountController struct {
	AccountService as.IAccountService
}

func NewAccountController(service as.IAccountService) IAccountController {
	return &AccountController{
		AccountService: service,
	}
}

func (ac AccountController) CreateNewAccount(c *fiber.Ctx) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Successfully create new account",
		"data":    []string{"Data Bank"},
	})

}
