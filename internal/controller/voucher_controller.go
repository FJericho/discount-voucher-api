package controller

import (
	"strconv"

	"github.com/FJericho/discount-voucher-api/internal/dto"
	"github.com/FJericho/discount-voucher-api/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type VoucherController interface {
	CreateVoucher(ctx *fiber.Ctx) error
	GetVouchers(ctx *fiber.Ctx) error
	GetVoucherById(ctx *fiber.Ctx) error
	UpdateVoucher(ctx *fiber.Ctx) error
	DeleteVoucher(ctx *fiber.Ctx) error
	UploadCSV(ctx *fiber.Ctx) error
	ExportCSV(ctx *fiber.Ctx) error
}

type VoucherControllerImpl struct {
	Log            *logrus.Logger
	VoucherService service.VoucherService
}

func NewVoucherController(log *logrus.Logger, voucherService service.VoucherService) VoucherController {
	return &VoucherControllerImpl{
		Log:            log,
		VoucherService: voucherService,
	}
}

func (c *VoucherControllerImpl) CreateVoucher(ctx *fiber.Ctx) error {
	var payload dto.VoucherRequest

	if err := ctx.BodyParser(&payload); err != nil {
		c.Log.Warnf("Failed to parse voucher create request: %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body.")
	}

	result, err := c.VoucherService.CreateVoucher(ctx.UserContext(), &payload)
	if err != nil {
		c.Log.Warnf("Failed to create voucher: %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(dto.WebResponse[*dto.VoucherResponse]{
		Message: "Voucher created successfully",
		Data:    result,
	})
}

func (c *VoucherControllerImpl) GetVouchers(ctx *fiber.Ctx) error {
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid page parameter. It must be a positive number.")
	}

	size, err := strconv.Atoi(ctx.Query("size", "10"))
	if err != nil || size < 1 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid size parameter. It must be a positive number.")
	}

	search := ctx.Query("search", "")
	order := ctx.Query("order", "asc")
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	result, meta, err := c.VoucherService.GetVouchers(ctx.UserContext(), page, size, search, order)
	if err != nil {
		c.Log.Warnf("Failed to get vouchers: %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[*[]dto.VoucherResponse]{
		Message: "Vouchers retrieved successfully.",
		Data:    result,
		Paging:  meta,
	})
}

func (c *VoucherControllerImpl) GetVoucherById(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Voucher ID is required.")
	}

	result, err := c.VoucherService.GetVoucherById(ctx.UserContext(), id)
	if err != nil {
		c.Log.Warnf("Failed to get voucher by ID: %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[*dto.VoucherResponse]{
		Message: "Voucher retrieved successfully",
		Data:    result,
	})
}

func (c *VoucherControllerImpl) UpdateVoucher(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Voucher ID is required.")
	}

	var payload dto.VoucherRequest
	if err := ctx.BodyParser(&payload); err != nil {
		c.Log.Warnf("Failed to parse voucher update request: %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body.")
	}

	result, err := c.VoucherService.UpdateVoucher(ctx.UserContext(), id, &payload)
	if err != nil {
		c.Log.Warnf("Failed to update voucher: %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[*dto.VoucherResponse]{
		Message: "Voucher updated successfully",
		Data:    result,
	})
}

func (c *VoucherControllerImpl) DeleteVoucher(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Voucher ID is required.")
	}

	err := c.VoucherService.DeleteVoucher(ctx.UserContext(), id)
	if err != nil {
		c.Log.Warnf("Failed to delete voucher: %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[any]{
		Message: "Voucher deleted successfully",
		Data:    nil,
	})
}

func (c *VoucherControllerImpl) UploadCSV(ctx *fiber.Ctx) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		c.Log.Warnf("Failed to get file from request: %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "CSV file is required.")
	}

	report, err := c.VoucherService.ImportFromCSV(ctx.UserContext(), file)
	if err != nil {
		c.Log.Errorf("CSV import failed: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to process CSV file.")
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[*dto.CSVImportReport]{
		Message: "CSV import completed.",
		Data:    report,
	})
}

func (c *VoucherControllerImpl) ExportCSV(ctx *fiber.Ctx) error {
	fileData, err := c.VoucherService.ExportToCSV(ctx.UserContext())
	if err != nil {
		c.Log.Errorf("Failed to export vouchers to CSV: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to export CSV")
	}

	ctx.Set("Content-Type", "text/csv")
	ctx.Set("Content-Disposition", `attachment; filename="vouchers.csv"`)
	return ctx.Send(fileData)
}

