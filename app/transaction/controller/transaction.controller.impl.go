package transaction_controller

import (
	ts "finance/app/transaction/service"
	v "finance/app/transaction/validation"
	g "finance/utility/generator"
	"strconv"

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

	resp, err := t.TransactionService.CreateNewTransaction(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Successfully create new transaction",
		"data":    resp,
	})
}

func (t TransactionController) UpdateTransaction(c *fiber.Ctx) error {
	req := new(v.UpdateTransactionRequest)
	err := c.BodyParser(req)
	if err != nil {
		log.Errorf("Failed to parse json: %v", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, "failed to parse request payload")
	}

	if lIdUser, ok := c.Locals("id_user").(int64); ok {
		req.IdUser = lIdUser
	}
	resp, err := t.TransactionService.UpdateTransaction(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Successfully update transaction",
		"data":    resp,
	})
}

func (t TransactionController) DeleteTransaction(c *fiber.Ctx) error {
	req := new(v.DeleteTransactionRequest)
	err := c.BodyParser(req)
	if err != nil {
		log.Errorf("Failed to parse json: %v", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, "failed to parse request payload")
	}

	if lIdUser, ok := c.Locals("id_user").(int64); ok {
		req.IdUser = lIdUser
	}

	err = t.TransactionService.DeleteTransaction(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Successfully delete transaction",
		"data":    "",
	})
}

func (t TransactionController) ListTransaction(c *fiber.Ctx) error {
	var userData v.UserTransactionDetailRequest

	idAccount := c.Query("id-account")
	userData.IdAccount = &idAccount
	filter := g.GenerateFilter(c)

	if lIdUser, ok := c.Locals("id_user").(int64); ok {
		sIdUser := strconv.Itoa(int(lIdUser))
		userData.IdUser = &sIdUser
	}

	resp, err := t.TransactionService.GetUserTransaction(filter, &userData)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Successfully delete transaction",
		"data":    resp,
	})
}

func (t TransactionController) CreateNewSubTransaction(c *fiber.Ctx) error {
	req := new(v.NewSubTransactionRequest)
	err := c.BodyParser(req)
	if err != nil {
		log.Errorf("Failed to parse json: %v", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, "failed to parse request payload")
	}

	if lIdUser, ok := c.Locals("id_user").(int64); ok {
		req.IdUser = lIdUser
	}

	resp, err := t.TransactionService.CreateNewSubTransaction(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Success! The detail-transaction has been added to your transaction.",
		"data":    resp,
	})
}
