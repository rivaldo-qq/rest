package service

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/entity"
	jwtentity "github.com/Dryluigi/go-grpc-ecommerce-be/internal/entity/jwt"
	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/repository"
	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/utils"
	"github.com/Dryluigi/go-grpc-ecommerce-be/pb/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	gocache "github.com/patrickmn/go-cache"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type IAuthService interface {
	Register(ctx context.Context, request *auth.RegisterRequest) (*auth.RegisterResponse, error)
	Login(ctx context.Context, request *auth.LoginRequest) (*auth.LoginResponse, error)
	Logout(ctx context.Context, request *auth.LogoutRequest) (*auth.LogoutResponse, error)
	ChangePassword(ctx context.Context, request *auth.ChangePasswordRequest) (*auth.ChangePasswordResponse, error)
	GetProfile(ctx context.Context, request *auth.GetProfileRequest) (*auth.GetProfileResponse, error)
}

type authService struct {
	authRepository repository.IAuthRepository
	cacheService   *gocache.Cache
}

func (as *authService) Register(ctx context.Context, request *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	if request.Password != request.PasswordConfirmation {
		return &auth.RegisterResponse{
			Base: utils.BadRequestResponse("Password is not matched"),
		}, nil
	}

	user, err := as.authRepository.GetUserByEmail(ctx, request.Email)
	if err != nil {
		return nil, err
	}

	if user != nil {
		return &auth.RegisterResponse{
			Base: utils.BadRequestResponse("User already exist"),
		}, nil
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), 10)
	if err != nil {
		return nil, err
	}

	// Insert ke db
	newUser := entity.User{
		Id:        uuid.NewString(),
		FullName:  request.FullName,
		Email:     request.Email,
		Password:  string(hashedPassword),
		RoleCode:  entity.UserRoleCustomer,
		CreatedAt: time.Now(),
		CreatedBy: &request.FullName,
	}
	err = as.authRepository.InsertUser(ctx, &newUser)
	if err != nil {
		return nil, err
	}

	return &auth.RegisterResponse{
		Base: utils.SuccessResponse("User is registered"),
	}, nil
}

func (as *authService) Login(ctx context.Context, request *auth.LoginRequest) (*auth.LoginResponse, error) {
	user, err := as.authRepository.GetUserByEmail(ctx, request.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &auth.LoginResponse{
			Base: utils.BadRequestResponse("User is not registered"),
		}, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated")
		}

		return nil, err
	}

	// generate jwt
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtentity.JwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Id,
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Email:    user.Email,
		FullName: user.FullName,
		Role:     user.RoleCode,
	})
	secretKey := os.Getenv("JWT_SECRET")
	accessToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	return &auth.LoginResponse{
		Base:        utils.SuccessResponse("Login successful"),
		AccessToken: accessToken,
	}, nil
}

func (as *authService) Logout(ctx context.Context, request *auth.LogoutRequest) (*auth.LogoutResponse, error) {
	jwtToken, err := jwtentity.ParseTokenFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tokenClaims, err := jwtentity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	as.cacheService.Set(jwtToken, "", time.Duration(tokenClaims.ExpiresAt.Time.Unix()-time.Now().Unix())*time.Second)

	return &auth.LogoutResponse{
		Base: utils.SuccessResponse("Logout success"),
	}, nil
}

func (as *authService) ChangePassword(ctx context.Context, request *auth.ChangePasswordRequest) (*auth.ChangePasswordResponse, error) {
	// Cek apakah new pass confirmation matched
	if request.NewPassword != request.NewPasswordConfirmation {
		return &auth.ChangePasswordResponse{
			Base: utils.BadRequestResponse("New password is not matched"),
		}, nil
	}

	// Cek apakah old password sama
	jwtToken, err := jwtentity.ParseTokenFromContext(ctx)
	if err != nil {
		return nil, err
	}
	claims, err := jwtentity.GetClaimsFromToken(jwtToken)
	if err != nil {
		return nil, err
	}

	user, err := as.authRepository.GetUserByEmail(ctx, claims.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &auth.ChangePasswordResponse{
			Base: utils.BadRequestResponse("User doesn't exist"),
		}, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.OldPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return &auth.ChangePasswordResponse{
				Base: utils.BadRequestResponse("Old password is not matched"),
			}, nil
		}

		return nil, err
	}

	// Update new password ke database
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), 10)
	if err != nil {
		return nil, err
	}
	err = as.authRepository.UpdateUserPassword(ctx, user.Id, string(hashedNewPassword), user.FullName)
	if err != nil {
		return nil, err
	}

	// Kirim response
	return &auth.ChangePasswordResponse{
		Base: utils.SuccessResponse("Change password success"),
	}, nil
}

func (as *authService) GetProfile(ctx context.Context, request *auth.GetProfileRequest) (*auth.GetProfileResponse, error) {
	claims, err := jwtentity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	user, err := as.authRepository.GetUserByEmail(ctx, claims.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &auth.GetProfileResponse{
			Base: utils.BadRequestResponse("User doesn't exist"),
		}, nil
	}

	return &auth.GetProfileResponse{
		Base:        utils.SuccessResponse("Get profile success"),
		UserId:      claims.Subject,
		FullName:    claims.FullName,
		Email:       claims.Email,
		RoleCode:    claims.Role,
		MemberSince: timestamppb.New(user.CreatedAt),
	}, nil
}

func NewAuthService(authRepository repository.IAuthRepository, cacheService *gocache.Cache) IAuthService {
	return &authService{
		authRepository: authRepository,
		cacheService:   cacheService,
	}
}
