package config

import (
	"github.com/FJericho/discount-voucher-api/internal/controller"
	"github.com/FJericho/discount-voucher-api/internal/middleware"
	"github.com/FJericho/discount-voucher-api/internal/repository"
	"github.com/FJericho/discount-voucher-api/internal/router"
	"github.com/FJericho/discount-voucher-api/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type AppConfig struct {
	DB       *gorm.DB
	App      *fiber.App
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *viper.Viper
}

func StartServer(config *AppConfig) {
	authRepository := repository.NewAuthRepository(config.DB)
	voucherRepository := repository.NewVoucherRepository(config.DB)

	authenticationMiddleware := middleware.NewAuthenticationMiddleware(config.Config)
	authorizationMiddleware := middleware.NewAuthorizationMiddleware(authenticationMiddleware)

	authService := service.NewAuthService(config.Log, config.Config, config.Validate, authRepository, authenticationMiddleware)
	voucherService := service.NewVoucherService(voucherRepository, config.Validate, config.Log)

	authController := controller.NewAuthController(config.Log, authService)
	voucherController := controller.NewVoucherController(config.Log, voucherService)

	routeConfig := router.RouteConfig{
		App:                      config.App,
		AuthController:           authController,
		AuthenticationMiddleware: authenticationMiddleware,
		AuthorizationMiddleware:  authorizationMiddleware,
		VoucherController:        voucherController,
	}

	routeConfig.Setup()
}
