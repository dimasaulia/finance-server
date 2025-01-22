package user

import "github.com/gofiber/fiber/v2"

type IUserController interface {
	ManualRegistration(c *fiber.Ctx) error
	ManualLogin(c *fiber.Ctx) error
	GoogleLogin(c *fiber.Ctx) error
	GoogleLoginCallback(c *fiber.Ctx) error
}
