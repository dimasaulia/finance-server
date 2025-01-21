package auth_middleware

import (
	"finance/provider/jwt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func LoginRequired(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accessToken := c.Get("authorization")

		if accessToken == "" {
			return fiber.NewError(fiber.StatusForbidden, "you are not authorize to access this menu")
		}

		data, err := jwt.VerifyJWT(accessToken)
		if err != nil {
			return fiber.NewError(fiber.StatusForbidden, "you are not authorize to access this menu")
		}

		type UserQueryResult struct {
			Email    string
			Fullname string
			Username string
			RoleName string
			Provider string
		}

		var existingUser UserQueryResult
		row := db.Raw(`SELECT u.email, u.fullname, u.username, u.provider, r.name role_name FROM "user" u JOIN "role" r ON r.id_role = u.id_role WHERE u.username = ?`, data.Username).Scan(&existingUser).RowsAffected
		if row == 0 {
			return fiber.NewError(fiber.StatusForbidden, "you are not authorize to access this menu")
		}

		c.Locals("username", existingUser.Username)
		c.Locals("fullname", existingUser.Fullname)
		c.Locals("email", existingUser.Email)
		c.Locals("role", existingUser.RoleName)
		return c.Next()
	}
}
