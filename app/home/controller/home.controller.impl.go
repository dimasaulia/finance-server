package controller

import "github.com/gofiber/fiber/v2"

type IHomeController interface {
	Home(c *fiber.Ctx) error
	HomePrivate(c *fiber.Ctx) error
}
