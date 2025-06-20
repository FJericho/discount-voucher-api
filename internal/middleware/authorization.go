package middleware

import (
	"github.com/FJericho/discount-voucher-api/internal/dto"
	"github.com/gofiber/fiber/v2"
)

type Authorization interface {
	AuthorizeAdmin(ctx *fiber.Ctx) error
}

type AuthorizationMiddleware struct {
	Authentication Authentication
}

func NewAuthorizationMiddleware(authentication Authentication) Authorization {
	return &AuthorizationMiddleware{
		Authentication: authentication,
	}
}

func (a *AuthorizationMiddleware) AuthorizeAdmin(ctx *fiber.Ctx) error {
	user := a.Authentication.GetCurrentUserAccount(ctx)

	if user.ID == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.WebResponse[any]{
			Errors: &dto.ErrorResponse{
				Message: "forbidden: unauthorized request, please login",
			},
		})
	}

	if user.Role != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(dto.WebResponse[any]{
			Errors: &dto.ErrorResponse{
				Message: "forbidden: admin access required",
			},
		})
	}

	return ctx.Next()
}
