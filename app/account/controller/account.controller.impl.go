package account_controller

import (
	as "finance/app/account/service"
	av "finance/app/account/validation"
	g "finance/utility/generator"
	"strconv"

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

func (ac AccountController) UserAccount(c *fiber.Ctx) error {
	data := av.AccountListRequest{}
	data.Type = c.Query("type")

	filterData := g.GenerateFilter(c)

	if lIdUserm, ok := c.Locals("id_user").(int64); ok {
		data.IdUser = strconv.Itoa(int(lIdUserm))
	}

	resp, err := ac.AccountService.UserAccountList(filterData, &data)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    resp,
		"message": "Successfully get all user account",
		"success": true,
	})
}

func (ac AccountController) DeleteAccount(c *fiber.Ctx) error {
	var idUser string
	var idAccount string = c.Params("id")

	if lIdUser, ok := c.Locals("id_user").(int64); ok {
		idUser = strconv.Itoa(int(lIdUser))
	}

	deleteAccountCount, err := ac.AccountService.DeleteAccountList(idAccount, idUser)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Successfully delete account",
		"data":    deleteAccountCount,
	})
}
