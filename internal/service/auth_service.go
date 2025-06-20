package service

import (
	"context"

	"github.com/FJericho/discount-voucher-api/internal/dto"
	"github.com/FJericho/discount-voucher-api/internal/entity"
	"github.com/FJericho/discount-voucher-api/internal/helper"
	"github.com/FJericho/discount-voucher-api/internal/middleware"
	"github.com/FJericho/discount-voucher-api/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type AuthService interface {
	Register(ctx context.Context, payload *dto.RegisterRequest) (*dto.AccountResponse, error)
	Login(ctx context.Context, payload *dto.LoginRequest) (*dto.LoginResponse, error)
}

type AuthServiceImpl struct {
	Log            *logrus.Logger
	Viper          *viper.Viper
	Validate       *validator.Validate
	AuthRepository repository.AuthRepository
	Authentication middleware.Authentication
}

func NewAuthService(log *logrus.Logger, viper *viper.Viper, validate *validator.Validate, authRepository repository.AuthRepository, authentication middleware.Authentication) AuthService {
	return &AuthServiceImpl{
		Log:            log,
		Viper:          viper,
		Validate:       validate,
		AuthRepository: authRepository,
		Authentication: authentication,
	}
}

func (s *AuthServiceImpl) Register(ctx context.Context, payload *dto.RegisterRequest) (*dto.AccountResponse, error) {
	err := s.Validate.Struct(payload)
	if err != nil {
		s.Log.Warnf("Invalid validation: %+v", err)
		errorResponse := helper.GenerateValidationErrors(err)
		return nil, fiber.NewError(fiber.StatusBadRequest, errorResponse)
	}

	exist, err := s.AuthRepository.CheckEmailIfExist(ctx, payload.Email)
	if exist {
		s.Log.Warn("Email already in use")
		return nil, fiber.NewError(fiber.StatusUnprocessableEntity, "Email already in use")
	}

	hashPassword, err := helper.HashPassword(payload.Password)
	if err != nil {
		s.Log.Warnf("Failed to generate bcrype hash : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to create hash password")
	}

	user, err := s.AuthRepository.CreateAccount(ctx, &entity.Account{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: hashPassword,
		Role:     entity.ADMIN,
	})

	if err != nil {
		s.Log.Warnf("Failed create user account to database : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to register a user account")
	}

	return &dto.AccountResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: &user.CreatedAt,
	}, nil
}

func (s *AuthServiceImpl) Login(ctx context.Context, payload *dto.LoginRequest) (*dto.LoginResponse, error) {
	err := s.Validate.Struct(payload)
	if err != nil {
		s.Log.Warnf("Invalid validation: %+v", err)
		errorResponse := helper.GenerateValidationErrors(err)
		return nil, fiber.NewError(fiber.StatusBadRequest, errorResponse)
	}

	exist, err := s.AuthRepository.CheckEmailIfExist(ctx, payload.Email)
	if !exist {
		s.Log.Warn("Email not found")
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Incorrect email or password")
	}

	user, _ := s.AuthRepository.FindUserByEmail(ctx, payload.Email)

	if err := helper.ComparePassword(user.Password, payload.Password); err != nil {
		s.Log.Warnf("Failed to compare user password with bcrype hash : %+v", err)
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Incorrect email or password")
	}

	token, err := s.Authentication.GenerateToken(user.ID, user.Email, user.Role, user.Name)
	if err != nil {
		s.Log.Warnf("Failed generated token : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed generated token")
	}

	return &dto.LoginResponse{
		Token: token,
	}, nil
}
