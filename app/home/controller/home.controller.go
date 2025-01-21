package controller

import "github.com/gofiber/fiber/v2"

type HomeController struct {
}

func NewHomeController() IHomeController {
	return &HomeController{}
}

func (h HomeController) Home(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"message": "Server Runing",
	})
}

func (h HomeController) HomePrivate(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"message": "Helo This Is Private Route",
	})
}
