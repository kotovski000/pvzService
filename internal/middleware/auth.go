package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"pvzService/internal/models"
	"strings"
)

func AuthMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.Error{Message: "Missing authorization header"})
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		claims := jwt.MapClaims{}
		tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(models.Error{Message: "Invalid token"})
		}

		if !tkn.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(models.Error{Message: "Invalid token"})
		}

		c.Locals("claims", claims)
		return c.Next()
	}
}

func CheckRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := c.Locals("claims").(jwt.MapClaims)

		role, ok := claims["role"].(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(models.Error{Message: "Invalid role in token"})
		}

		for _, allowedRole := range roles {
			if role == allowedRole {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(models.Error{Message: "Insufficient role"})
	}
}
