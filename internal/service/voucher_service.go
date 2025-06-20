package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/FJericho/discount-voucher-api/internal/dto"
	"github.com/FJericho/discount-voucher-api/internal/entity"
	"github.com/FJericho/discount-voucher-api/internal/helper"
	"github.com/FJericho/discount-voucher-api/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type VoucherService interface {
	CreateVoucher(ctx context.Context, a *dto.VoucherRequest) (*dto.VoucherResponse, error)
	GetVouchers(ctx context.Context, page, size int, search, order string) (*[]dto.VoucherResponse, *dto.PageMetadata, error)
	GetVoucherById(ctx context.Context, id string) (*dto.VoucherResponse, error)
	DeleteVoucher(ctx context.Context, id string) error
	UpdateVoucher(ctx context.Context, id string, req *dto.VoucherRequest) (*dto.VoucherResponse, error)
	ImportFromCSV(ctx context.Context, file *multipart.FileHeader) (*dto.CSVImportReport, error)
	ExportToCSV(ctx context.Context) ([]byte, error)
}

type VoucherServiceImpl struct {
	VoucherRepository repository.VoucherRepository
	Validate          *validator.Validate
	Log               *logrus.Logger
}

func NewVoucherService(repo repository.VoucherRepository, validate *validator.Validate, log *logrus.Logger) VoucherService {
	return &VoucherServiceImpl{
		VoucherRepository: repo,
		Validate:          validate,
		Log:               log,
	}
}

func (s *VoucherServiceImpl) CreateVoucher(ctx context.Context, a *dto.VoucherRequest) (*dto.VoucherResponse, error) {
	err := s.Validate.Struct(a)
	if err != nil {
		s.Log.Warnf("Invalid validation: %+v", err)
		errorResponse := helper.GenerateValidationErrors(err)
		return nil, fiber.NewError(fiber.StatusBadRequest, errorResponse)
	}

	exists, err := s.VoucherRepository.CheckVoucherIfExist(ctx, a.Code)
	if err != nil {
		s.Log.Errorf("Failed to check existing voucher: %v", err)
		return nil, err
	}
	if exists {
		s.Log.Warnf("Voucher code already exists: %s", a.Code)
		return nil, fiber.NewError(fiber.StatusConflict, "voucher code already exists")
	}

	newVoucher := &entity.Voucher{
		Code:            a.Code,
		DiscountPercent: a.DiscountPercent,
		ExpiryDate:      a.ExpiryDate,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	res, err := s.VoucherRepository.CreateVoucher(ctx, newVoucher)
	if err != nil {
		s.Log.Errorf("Failed to create voucher: %v", err)
		return nil, err
	}

	s.Log.Infof("Voucher created successfully: ID=%s, Code=%s", res.ID, res.Code)

	return &dto.VoucherResponse{
		ID:              res.ID,
		Code:            res.Code,
		DiscountPercent: res.DiscountPercent,
		ExpiryDate:      res.ExpiryDate,
		CreatedAt:       &res.CreatedAt,
		UpdatedAt:       &res.UpdatedAt,
	}, nil
}

func (s *VoucherServiceImpl) GetVouchers(ctx context.Context, page, size int, search, order string) (*[]dto.VoucherResponse, *dto.PageMetadata, error) {
	offset := (page - 1) * size

	vouchers, total, err := s.VoucherRepository.GetVouchers(ctx, size, offset, search, order)
	if err != nil {
		s.Log.Errorf("Failed to retrieve vouchers: %v", err)
		return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve vouchers.")
	}

	var responses []dto.VoucherResponse
	for _, v := range vouchers {
		responses = append(responses, dto.VoucherResponse{
			ID:              v.ID,
			Code:            v.Code,
			DiscountPercent: v.DiscountPercent,
			ExpiryDate:      v.ExpiryDate,
			CreatedAt:       &v.CreatedAt,
			UpdatedAt:       &v.UpdatedAt,
		})
	}

	totalPages := int((total + int64(size) - 1) / int64(size))
	meta := &dto.PageMetadata{
		Page:        page,
		Size:        size,
		TotalItem:   total,
		TotalPage:   int64(totalPages),
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}

	return &responses, meta, nil
}

func (s *VoucherServiceImpl) GetVoucherById(ctx context.Context, id string) (*dto.VoucherResponse, error) {
	v, err := s.VoucherRepository.GetVoucherById(ctx, id)
	if err != nil {
		s.Log.Errorf("Failed to get voucher by ID %s: %v", id, err)
		return nil, err
	}

	s.Log.Infof("Voucher found: ID=%s, Code=%s", v.ID, v.Code)

	return &dto.VoucherResponse{
		ID:              v.ID,
		Code:            v.Code,
		DiscountPercent: v.DiscountPercent,
		ExpiryDate:      v.ExpiryDate,
		CreatedAt:       &v.CreatedAt,
		UpdatedAt:       &v.UpdatedAt,
	}, nil
}

func (s *VoucherServiceImpl) DeleteVoucher(ctx context.Context, id string) error {
	err := s.VoucherRepository.DeleteVoucher(ctx, id)
	if err != nil {
		s.Log.Errorf("Failed to delete voucher ID %s: %v", id, err)
		return err
	}

	s.Log.Infof("Voucher deleted: ID=%s", id)
	return nil
}

func (s *VoucherServiceImpl) UpdateVoucher(ctx context.Context, id string, req *dto.VoucherRequest) (*dto.VoucherResponse, error) {
	if err := s.Validate.Struct(req); err != nil {
		s.Log.Warnf("Validation failed for UpdateVoucher: %+v", err)
		return nil, fiber.NewError(fiber.StatusBadRequest, "validation error")
	}

	existing, err := s.VoucherRepository.GetVoucherById(ctx, id)
	if err != nil {
		s.Log.Errorf("Failed to find voucher for update: ID=%s: %v", id, err)
		return nil, err
	}

	if !strings.EqualFold(existing.Code, req.Code) {
		exists, err := s.VoucherRepository.CheckVoucherIfExist(ctx, req.Code)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fiber.NewError(fiber.StatusConflict, "voucher code already exists")
		}
	}

	existing.Code = req.Code
	existing.DiscountPercent = req.DiscountPercent
	existing.ExpiryDate = req.ExpiryDate
	existing.UpdatedAt = time.Now()

	updated, err := s.VoucherRepository.UpdateVoucher(ctx, existing)
	if err != nil {
		s.Log.Errorf("Failed to update voucher ID=%s: %v", id, err)
		return nil, err
	}

	s.Log.Infof("Voucher updated successfully: ID=%s", updated.ID)

	return &dto.VoucherResponse{
		ID:              updated.ID,
		Code:            updated.Code,
		DiscountPercent: updated.DiscountPercent,
		ExpiryDate:      updated.ExpiryDate,
		CreatedAt:       &updated.CreatedAt,
		UpdatedAt:       &updated.UpdatedAt,
	}, nil
}

func (s *VoucherServiceImpl) ImportFromCSV(ctx context.Context, file *multipart.FileHeader) (*dto.CSVImportReport, error) {
	f, err := file.Open()
	if err != nil {
		s.Log.Errorf("Failed to open uploaded CSV file: %v", err)
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.TrimLeadingSpace = true
	reader.ReuseRecord = true

	var successCount, failedCount int
	var failedRows []dto.CSVFailedRow

	records, err := reader.ReadAll()
	if err != nil {
		s.Log.Errorf("Failed to read CSV file: %v", err)
		return nil, err
	}

	for i, record := range records[1:] {
		if len(record) < 3 {
			failedCount++
			failedRows = append(failedRows, dto.CSVFailedRow{Row: i + 2, Reason: "Incomplete data"})
			continue
		}

		discount, err := strconv.Atoi(record[1])
		if err != nil || discount < 1 || discount > 100 {
			failedCount++
			failedRows = append(failedRows, dto.CSVFailedRow{Row: i + 2, Reason: "Invalid discount percent"})
			continue
		}

		expiry, err := time.Parse("2006-01-02", record[2])
		if err != nil {
			failedCount++
			failedRows = append(failedRows, dto.CSVFailedRow{Row: i + 2, Reason: "Invalid expiry date format"})
			continue
		}

		voucher := &entity.Voucher{
			Code:            record[0],
			DiscountPercent: discount,
			ExpiryDate:      expiry,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		_, err = s.VoucherRepository.CreateVoucher(ctx, voucher)
		if err != nil {
			s.Log.Warnf("Failed to insert voucher (row %d): %v", i+2, err)
			failedCount++
			failedRows = append(failedRows, dto.CSVFailedRow{Row: i + 2, Reason: "Database error or duplicate code"})
			continue
		}

		successCount++
	}

	s.Log.Infof("CSV Import completed: %d success, %d failed", successCount, failedCount)

	return &dto.CSVImportReport{
		SuccessCount: successCount,
		FailedCount:  failedCount,
		FailedRows:   failedRows,
	}, nil
}

func (s *VoucherServiceImpl) ExportToCSV(ctx context.Context) ([]byte, error) {
	vouchers, err := s.VoucherRepository.FindAll(ctx, "", "asc")
	if err != nil {
		s.Log.Errorf("Failed to fetch vouchers for export: %v", err)
		return nil, err
	}

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	_ = writer.Write([]string{"voucher_code", "discount_percent", "expiry_date"})

	for _, v := range vouchers {
		writer.Write([]string{
			v.Code,
			strconv.Itoa(v.DiscountPercent),
			v.ExpiryDate.Format("2006-01-02"),
		})
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		s.Log.Errorf("CSV writing error: %v", err)
		return nil, err
	}

	return buffer.Bytes(), nil
}
