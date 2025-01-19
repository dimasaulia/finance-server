package user

import "github.com/gofiber/fiber/v2"

type IUserController interface {
	ManualRegistration(c *fiber.Ctx) error
}
