package transaction_controller

import "github.com/gofiber/fiber/v2"

type ITransactionController interface {
	CreateNewTransaction(c *fiber.Ctx) error
	UpdateTransaction(c *fiber.Ctx) error
	DeleteTransaction(c *fiber.Ctx) error
	ListTransaction(c *fiber.Ctx) error
	CreateNewSubTransaction(c *fiber.Ctx) error
	UpdateSubTransaction(c *fiber.Ctx) error
	DeleteSubTransaction(c *fiber.Ctx) error
}
