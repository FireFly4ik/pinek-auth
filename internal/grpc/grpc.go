package grpc

import (
	"auth/internal/config"
	pb "auth/internal/proto"
	"context"
	"gorm.io/gorm"
)

type AuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
	database *gorm.DB
	envConf  *config.Config
}

func NewAuthServer(db *gorm.DB, cfg *config.Config) *AuthServiceServer {
	return &AuthServiceServer{
		database: db,
		envConf:  cfg,
	}
}

func (s *AuthServiceServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {

	return &pb.AuthResponse{Message: "User registered successfully"}, nil
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
