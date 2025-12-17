package utils

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// ParamsInt extracts an integer parameter from the URL path
func ParamsInt(c fiber.Ctx, key string) (int, error) {
	param := c.Params(key)
	return strconv.Atoi(param)
}
