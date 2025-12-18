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

// QueryInt extracts an integer query parameter with a default value
func QueryInt(c fiber.Ctx, key string, defaultValue int) int {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}
