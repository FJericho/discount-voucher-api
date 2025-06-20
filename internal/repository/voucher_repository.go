package repository

import (
	"context"
	"strings"

	"github.com/FJericho/discount-voucher-api/internal/entity"
	"gorm.io/gorm"
)

type VoucherRepository interface {
	CreateVoucher(ctx context.Context, a *entity.Voucher) (*entity.Voucher, error)
	CheckVoucherIfExist(ctx context.Context, title string) (bool, error)
	GetVouchers(ctx context.Context, size, offset int, search, order string) ([]*entity.Voucher, int64, error)
	GetVoucherById(ctx context.Context, id string) (*entity.Voucher, error)
	DeleteVoucher(ctx context.Context, id string) error
	UpdateVoucher(ctx context.Context, a *entity.Voucher) (*entity.Voucher, error)
	FindAll(ctx context.Context, search, order string) ([]*entity.Voucher, error)

}

type VoucherRepositoryImpl struct {
	DB *gorm.DB
}

func NewVoucherRepository(db *gorm.DB) VoucherRepository {
	return &VoucherRepositoryImpl{
		DB: db,
	}
}

func (v *VoucherRepositoryImpl) CreateVoucher(ctx context.Context, a *entity.Voucher) (*entity.Voucher, error) {
	if err := v.DB.WithContext(ctx).Create(a).Error; err != nil {
		return nil, err
	}
	return a, nil
}

func (v *VoucherRepositoryImpl) CheckVoucherIfExist(ctx context.Context, code string) (bool, error) {
	var voucher entity.Voucher

	result := v.DB.WithContext(ctx).Where("code = ?", code).First(&voucher)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, result.Error
	}
	return true, nil
}

func (v *VoucherRepositoryImpl) GetVoucherById(ctx context.Context, id string) (*entity.Voucher, error) {
	var voucher entity.Voucher
	if err := v.DB.WithContext(ctx).First(&voucher, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &voucher, nil
}

func (r *VoucherRepositoryImpl) GetVouchers(ctx context.Context, size, offset int, search, order string) ([]*entity.Voucher, int64, error) {
	var vouchers []*entity.Voucher
	var total int64

	query := r.DB.WithContext(ctx).Model(&entity.Voucher{})

	if search != "" {
		query = query.Where("LOWER(code) LIKE ?", "%"+strings.ToLower(search)+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if order == "asc" {
		query = query.Order("created_at asc")
	} else {
		query = query.Order("created_at desc")
	}

	if err := query.Limit(size).Offset(offset).Find(&vouchers).Error; err != nil {
		return nil, 0, err
	}

	return vouchers, total, nil
}

func (v *VoucherRepositoryImpl) DeleteVoucher(ctx context.Context, id string) error {
	if err := v.DB.WithContext(ctx).Where("id = ?", id).Delete(&entity.Voucher{}).Error; err != nil {
		return err
	}
	return nil
}

func (v *VoucherRepositoryImpl) UpdateVoucher(ctx context.Context, updated *entity.Voucher) (*entity.Voucher, error) {
	var existing entity.Voucher

	if err := v.DB.WithContext(ctx).First(&existing, "id = ?", updated.ID).Error; err != nil {
		return nil, err
	}

	existing.Code = updated.Code
	existing.DiscountPercent = updated.DiscountPercent
	existing.ExpiryDate = updated.ExpiryDate

	if err := v.DB.WithContext(ctx).Save(&existing).Error; err != nil {
		return nil, err
	}

	return &existing, nil
}

func (r *VoucherRepositoryImpl) FindAll(ctx context.Context, search, order string) ([]*entity.Voucher, error) {
	var vouchers []*entity.Voucher
	query := r.DB.WithContext(ctx).Model(&entity.Voucher{})

	if search != "" {
		query = query.Where("LOWER(code) LIKE ?", "%"+strings.ToLower(search)+"%")
	}

	if order == "asc" {
		query = query.Order("created_at ASC")
	} else {
		query = query.Order("created_at DESC")
	}

	if err := query.Find(&vouchers).Error; err != nil {
		return nil, err
	}

	return vouchers, nil
}


