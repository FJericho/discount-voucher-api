package router

import (
	"github.com/FJericho/discount-voucher-api/internal/controller"
	"github.com/FJericho/discount-voucher-api/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

type RouteConfig struct {
	App                      *fiber.App
	AuthController           controller.AuthController
	VoucherController        controller.VoucherController
	AuthenticationMiddleware middleware.Authentication
	AuthorizationMiddleware  middleware.Authorization
}

func (r *RouteConfig) Setup() {
	r.SetupPublicRoute()
	r.SetupAdminRoute()
}

func (r *RouteConfig) SetupPublicRoute() {
	r.App.Post("/api/v1/login", r.AuthController.Login)
	r.App.Post("/api/v1/register", r.AuthController.Register)

}

func (r *RouteConfig) SetupAdminRoute() {
	admin := r.App.Group("/api/v1", r.AuthenticationMiddleware.Authorize, r.AuthorizationMiddleware.AuthorizeAdmin)

	voucher := admin.Group("/voucher")
	voucher.Post("/", r.VoucherController.CreateVoucher)
	voucher.Get("/", r.VoucherController.GetVouchers)
	voucher.Get("/export", r.VoucherController.ExportCSV)
	voucher.Post("/upload-csv", r.VoucherController.UploadCSV)
	voucher.Get("/:id", r.VoucherController.GetVoucherById)
	voucher.Patch("/:id", r.VoucherController.UpdateVoucher)
	voucher.Delete("/:id", r.VoucherController.DeleteVoucher)
}
