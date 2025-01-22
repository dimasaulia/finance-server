package user

import (
	s "finance/app/user/service"
	v "finance/app/user/validation"
	"fmt"
	"math/rand"
	"strings"
	"time"

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

func (h UserController) ManualLogin(c *fiber.Ctx) error {
	req := new(v.UserLoginRequest)
	err := c.BodyParser(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	req.Provider = "MANUAL"
	resp, err := h.Service.UserLogin(*req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "Success Logingin User",
		"data":    resp,
	})
}

func (h UserController) GoogleLogin(c *fiber.Ctx) error {

	resp, err := h.Service.GenerateGoogleLoginUrl()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	c.Cookie(&fiber.Cookie{
		Name:     "state",
		Value:    resp.State,
		Expires:  time.Now().Add(time.Minute * 2),
		HTTPOnly: true,
	})

	return c.Status(fiber.StatusTemporaryRedirect).Redirect(resp.RedirectUrl)
}

func (h UserController) GoogleLoginCallback(c *fiber.Ctx) error {
	req := c.Queries()
	state := req["state"]
	code := req["code"]

	cookieState := c.Cookies("state")

	if cookieState != state {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid state token")
	}

	userInfo, err := h.Service.GoogleLoginCallback(code)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "failed to login with google")
	}

	// Reset state cookie
	c.Cookie(&fiber.Cookie{
		Name:     "state",
		Value:    "",
		Expires:  time.Now().Add(time.Second + 1),
		HTTPOnly: true,
	})

	register, err := h.Service.UserRegistartion(v.UserRegistrationRequest{
		Email:      userInfo.Email,
		Fullname:   fmt.Sprintf("%s %s", userInfo.GivenName, userInfo.FamilyName),
		Username:   fmt.Sprintf("%s_%v", strings.ReplaceAll(userInfo.GivenName, " ", "_"), rand.Intn(999-100+1)+100),
		Provider:   "GOOGLE",
		ProviderId: userInfo.Sub,
	})

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Failed to register new user with google")
	}

	data, err := h.Service.UserLogin(v.UserLoginRequest{
		UsernameOrEmail: register.Email,
		Provider:        "GOOGLE",
		ProviderId:      userInfo.Sub,
	})

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusTemporaryRedirect).JSON(fiber.Map{
		"message": "Success login with user",
		"data":    data,
	})
}
