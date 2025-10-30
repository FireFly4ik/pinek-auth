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

	userID, err := s.database.RegisterUser(req.Username, req.Login, hashedPassword)
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

	refreshToken, err := s.jwtService.GenerateRefreshToken(userID)
	if err != nil {
		_ = s.database.DeleteUser(userID)
		log.Error().Err(err).Msg("failed to generate refresh token")
		return &pb.AuthResponse{Message: "Internal server error"}, err
	}

	hashedRefreshToken := pkg.HashRefreshToken(refreshToken)

	err = s.database.StoreRefreshToken(userID, []byte(hashedRefreshToken))
	if err != nil {
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

	return &pb.AuthResponse{Message: "User logged in successfully"}, nil
}

func (s *AuthServiceServer) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.AuthResponse, error) {

	return &pb.AuthResponse{Message: "Token refreshed successfully"}, nil
}

func (s *AuthServiceServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {

	return &pb.LogoutResponse{Message: "User logged out successfully"}, nil
}
