package controller

import (
	"fmt"

	"github.com/FJericho/discount-voucher-api/internal/dto"
	"github.com/FJericho/discount-voucher-api/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type AuthController interface {
	Register(ctx *fiber.Ctx) error
	Login(ctx *fiber.Ctx) error
}

type AuthControllerImpl struct {
	Log         *logrus.Logger
	AuthService service.AuthService
}

func NewAuthController(log *logrus.Logger, authService service.AuthService) AuthController {
	return &AuthControllerImpl{
		Log:         log,
		AuthService: authService,
	}
}

func (c *AuthControllerImpl) Register(ctx *fiber.Ctx) error {
	var payload dto.RegisterRequest

	if err := ctx.BodyParser(&payload); err != nil {
		c.Log.Warnf("Failed to parse request body : %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body.")
	}

	response, err := c.AuthService.Register(ctx.UserContext(), &payload)
	if err != nil {
		c.Log.Warnf("Failed to register user : %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(dto.WebResponse[*dto.AccountResponse]{
		Message: "Register Successfully",
		Data:    response,
	})
}

func (c *AuthControllerImpl) Login(ctx *fiber.Ctx) error {
	var payload dto.LoginRequest

	fmt.Println(payload)

	if err := ctx.BodyParser(&payload); err != nil {
		c.Log.Warnf("Failed to parse request body : %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body.")
	}

	response, err := c.AuthService.Login(ctx.UserContext(), &payload)

	if err != nil {
		c.Log.Warnf("Failed to login user : %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[*dto.LoginResponse]{
		Data:    response,
		Message: "Login Successfully",
	})
}
