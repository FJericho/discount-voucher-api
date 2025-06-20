package repository

import (
	"context"

	"github.com/FJericho/discount-voucher-api/internal/entity"
	"gorm.io/gorm"
)

type AuthRepository interface {
	CreateAccount(ctx context.Context, a *entity.Account) (*entity.Account, error)
	FindUserByEmail(ctx context.Context, email string) (*entity.Account, error)
	CheckEmailIfExist(ctx context.Context, email string) (bool, error)
}

type AuthRepositoryImpl struct {
	DB *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &AuthRepositoryImpl{
		DB: db,
	}
}

func (r *AuthRepositoryImpl) CreateAccount(ctx context.Context, a *entity.Account) (*entity.Account, error) {
	if err := r.DB.WithContext(ctx).Create(a).Error; err != nil {
		return nil, err
	}
	return a, nil
}

func (r *AuthRepositoryImpl) CheckEmailIfExist(ctx context.Context, email string) (bool, error) {
	var account entity.Account

	result := r.DB.WithContext(ctx).Where("email = ?", email).First(&account)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, result.Error
	}
	return true, nil
}

func (r *AuthRepositoryImpl) FindUserByEmail(ctx context.Context, email string) (*entity.Account, error) {
	var account entity.Account

	err := r.DB.WithContext(ctx).Where("email = ?", email).First(&account).Error

	if err != nil {
		return nil, err
	}

	return &account, nil
}
