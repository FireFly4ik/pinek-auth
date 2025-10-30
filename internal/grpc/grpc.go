package grpc

import (
	"auth/internal/config"
	"auth/internal/db"
	"auth/internal/jwt"
	pb "auth/internal/proto"
	"auth/pkg"
	"context"
	"errors"
	"github.com/rs/zerolog/log"
)

type AuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
	database   *db.AuthDatabase
	envConf    *config.Config
	jwtService jwt.JWTService
}

func NewAuthServer(db *db.AuthDatabase, cfg *config.Config, jwtService jwt.JWTService) *AuthServiceServer {
	return &AuthServiceServer{
		database:   db,
		envConf:    cfg,
		jwtService: jwtService,
	}
}

func (s *AuthServiceServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	hashedPassword, err := pkg.HashPassword(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("failed to hash password")
		return &pb.AuthResponse{Message: "Internal server error"}, err
	}

	userID, err := s.database.RegisterUser(req.Username, req.Login, hashedPassword, "user")
	if err != nil {
		if errors.Is(err, db.ErrUserExists) {
			return &pb.AuthResponse{Message: "User already exists"}, nil
		}
		log.Error().Err(err).Msg("failed to register user")
		return &pb.AuthResponse{Message: "Internal server error"}, err
	}

	accessToken, err := s.jwtService.GenerateAccessToken(userID, req.Username, "user")
	if err != nil {
		_ = s.database.DeleteUser(userID)
		log.Error().Err(err).Msg("failed to generate access token")
		return &pb.AuthResponse{Message: "Internal server error"}, err
	}

	expiresAt, refreshToken, err := s.jwtService.GenerateRefreshToken(userID)
	if err != nil {
		_ = s.database.DeleteUser(userID)
		log.Error().Err(err).Msg("failed to generate refresh token")
		return &pb.AuthResponse{Message: "Internal server error"}, err
	}

	hashedRefreshToken := pkg.HashRefreshToken(refreshToken)

	if err := s.database.StoreRefreshToken(userID, []byte(hashedRefreshToken), expiresAt); err != nil {
		_ = s.database.DeleteUser(userID)
		log.Error().Err(err).Msg("failed to store refresh token")
		return &pb.AuthResponse{Message: "Internal server error"}, err
	}

	return &pb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Message:      "User registered successfully",
	}, nil
}

func (s *AuthServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	userID, username, hashedPassword, role, err := s.database.AuthenticateUser(req.Login)
	if err != nil {
		if errors.Is(err, db.ErrInvalidCredentials) {
			return &pb.AuthResponse{Message: "Invalid credentials"}, nil
		}
		log.Error().Err(err).Msg("failed to authenticate user")
		return &pb.AuthResponse{Message: "Internal server error"}, err
	}

	if !pkg.CheckPasswordHash(req.Password, hashedPassword) {
		return &pb.AuthResponse{Message: "Invalid credentials"}, nil
	}

	accessToken, err := s.jwtService.GenerateAccessToken(userID, username, role)
	if err != nil {
		_ = s.database.DeleteUser(userID)
		log.Error().Err(err).Msg("failed to generate access token")
		return &pb.AuthResponse{Message: "Internal server error"}, err
	}

	expiresAt, refreshToken, err := s.jwtService.GenerateRefreshToken(userID)
	if err != nil {
		_ = s.database.DeleteUser(userID)
		log.Error().Err(err).Msg("failed to generate refresh token")
		return &pb.AuthResponse{Message: "Internal server error"}, err
	}

	hashedRefreshToken := pkg.HashRefreshToken(refreshToken)

	if err := s.database.StoreRefreshToken(userID, []byte(hashedRefreshToken), expiresAt); err != nil {
		_ = s.database.DeleteUser(userID)
		log.Error().Err(err).Msg("failed to store refresh token")
		return &pb.AuthResponse{Message: "Internal server error"}, err
	}

	return &pb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Message:      "User logged successfully",
	}, nil
}

func (s *AuthServiceServer) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.AuthResponse, error) {
	claims, err := s.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return &pb.AuthResponse{Message: "Invalid refresh token"}, nil
	}

	usedTokenHash := pkg.HashRefreshToken(req.RefreshToken)

	err = s.database.CheckRefreshTokenHash(claims.UserID, []byte(usedTokenHash))
	if err != nil {
		if errors.Is(err, db.ErrTokenNotFound) {
			return &pb.AuthResponse{Message: "Invalid refresh token"}, nil
		}
		log.Error().Err(err).Msg("failed to get refresh token hash")
		return &pb.AuthResponse{Message: "Internal server error"}, err
	}

	registration, err := s.database.GetRegistrationByID(claims.UserID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user registration")
		return &pb.AuthResponse{Message: "Internal server error"}, err
	}

	accessToken, err := s.jwtService.GenerateAccessToken(claims.UserID, registration.Username, registration.Role)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate access token")
		return &pb.AuthResponse{Message: "Internal server error"}, err
	}

	expiresAt, refreshToken, err := s.jwtService.GenerateRefreshToken(claims.UserID)
	if err != nil {
		_ = s.database.DeleteUser(claims.UserID)
		log.Error().Err(err).Msg("failed to generate refresh token")
		return &pb.AuthResponse{Message: "Internal server error"}, err
	}

	hashedRefreshToken := pkg.HashRefreshToken(refreshToken)

	if err := s.database.StoreRefreshToken(claims.UserID, []byte(hashedRefreshToken), expiresAt); err != nil {
		log.Error().Err(err).Msg("failed to store refresh token")
		return &pb.AuthResponse{Message: "Internal server error"}, err
	}

	if err := s.database.RevokeRefreshToken([]byte(usedTokenHash), "RefreshToken used"); err != nil {
		_ = s.database.RevokeRefreshToken([]byte(hashedRefreshToken), "New refresh token revoked due to old token revoke failure")
		log.Error().Err(err).Msg("failed to revoke old refresh token")
		return &pb.AuthResponse{Message: "Internal server error"}, err
	}

	return &pb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Message:      "Token refreshed successfully",
	}, nil
}

func (s *AuthServiceServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	claims, err := s.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return &pb.LogoutResponse{Message: "Invalid refresh token"}, nil
	}

	usedTokenHash := pkg.HashRefreshToken(req.RefreshToken)

	err = s.database.CheckRefreshTokenHash(claims.UserID, []byte(usedTokenHash))
	if err != nil {
		if errors.Is(err, db.ErrTokenNotFound) {
			return &pb.LogoutResponse{Message: "Invalid refresh token"}, nil
		}
		log.Error().Err(err).Msg("failed to get refresh token hash")
		return &pb.LogoutResponse{Message: "Internal server error"}, err
	}

	if err := s.database.RevokeRefreshToken([]byte(usedTokenHash), "Logout"); err != nil {
		log.Error().Err(err).Msg("failed to revoke old refresh token")
		return &pb.LogoutResponse{Message: "Internal server error"}, err
	}

	return &pb.LogoutResponse{Message: "User logged out successfully"}, nil
}
