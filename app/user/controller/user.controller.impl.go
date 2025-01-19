package user

import (
	s "finance/app/user/service"
	v "finance/app/user/validation"

	"github.com/gofiber/fiber/v2"
)

type UserController struct {
	Service s.IUserService
}

func NewUserController(s s.IUserService) IUserController {
	return &UserController{
		Service: s,
	}
}

func (h UserController) ManualRegistration(c *fiber.Ctx) error {
	req := new(v.UserRegistrationRequest)
	err := c.BodyParser(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	req.Provider = "MANUAL"
	resp, err := h.Service.UserRegistartion(*req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "Success Create New User",
		"data":    resp,
	})
}
