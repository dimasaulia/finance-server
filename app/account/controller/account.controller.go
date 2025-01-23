package account_controller

import "github.com/gofiber/fiber/v2"

type IAccountController interface {
	CreateNewAccount(c *fiber.Ctx) error
}
