package transaction_controller

import (
	ts "finance/app/transaction/service"
	v "finance/app/transaction/validation"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type TransactionController struct {
	TransactionService ts.ITransactionService
}

func NewTransactionController(s ts.ITransactionService) ITransactionController {
	return &TransactionController{
		TransactionService: s,
	}
}

func (t TransactionController) CreateNewTransaction(c *fiber.Ctx) error {
	req := new(v.NewTransactionRequest)
	err := c.BodyParser(req)
	if err != nil {
		log.Errorf("Failed to parse json: %v", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, "failed to parse request payload")
	}

	if lIdUser, ok := c.Locals("id_user").(int64); ok {
		req.IdUser = lIdUser
	}

	_, err = t.TransactionService.CreateNewTransaction(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Successfullt create new transaction",
		"data":    []string{},
	})
}
