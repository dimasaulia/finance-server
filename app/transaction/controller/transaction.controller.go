package transaction_controller

import "github.com/gofiber/fiber/v2"

type ITransactionController interface {
	CreateNewTransaction(c *fiber.Ctx) error
}
