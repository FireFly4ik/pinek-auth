package main

import (
	"auth/internal/config"
	"auth/internal/consul"
	"auth/internal/db"
	grpcService "auth/internal/grpc"
	"auth/internal/logger"
	pb "auth/internal/proto"
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("No .env file found")
	}
	envConf := config.NewEnvConfig()
	config.PrintConfigWithHiddenSecrets(envConf)

	logger.Setup(envConf.ProductionType)

	database := db.ConnectDB(envConf)

	consulProvider := consul.NewProvider(envConf)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", envConf.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen: %v")
	}

	server := grpcService.NewAuthServer(database, envConf)

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, server)

	go func() {
		log.Info().Msgf("auth service listening on port %s", envConf.Port)
		if err := grpcServer.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatal().Err(err).Msg("failed to serve")
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case s := <-sig:
		log.Info().Msg(fmt.Sprintf("signal received: %s — starting graceful shutdown", s))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-ctx.Done():
		log.Warn().Msg("graceful shutdown timeout — forcing stop")
		grpcServer.Stop()
	case <-done:
		log.Info().Msg("gRPC server stopped gracefully")
	}

	consulProvider.DeregisterService()

	log.Info().Msg("auth service shutdown gracefully")
}
