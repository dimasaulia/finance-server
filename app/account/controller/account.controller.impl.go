package account_controller

import (
	as "finance/app/account/service"
	av "finance/app/account/validation"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
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
	req := new(av.AccountCreationRequest)
	err := c.BodyParser(req)
	if err != nil {
		log.Error(err)
		return fiber.NewError(fiber.StatusBadRequest, "failed to parse request payload")
	}

	if lIdUserm, ok := c.Locals("id_user").(int64); ok {
		req.IdUser = lIdUserm
	}

	resp, err := ac.AccountService.CreateAccount(*req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Successfully create new account",
		"data":    resp,
	})

}
