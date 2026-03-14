package main

import (
	"banking-service/internal/client"
	"banking-service/internal/config"
	"banking-service/internal/handler"
	"banking-service/internal/permission"
	"banking-service/internal/server"
	"common/pkg/auth"
	"common/pkg/jwt"
	"common/pkg/logging"
	"common/pkg/pb"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

// @title Banking Service API
// @version 1.0
// @description API for managing accounts, balances, and banking operations.
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme.
func main() {
	fx.New(
		fx.Provide(
			config.Load,
			func(cfg *config.Configuration) auth.TokenVerifier {
				return jwt.NewJWTVerifier(cfg.JWTSecret)
			},
			client.NewUserServiceClient,
			func(conn *grpc.ClientConn) pb.PermissionServiceClient {
				return pb.NewPermissionServiceClient(conn)
			},
			func(client pb.PermissionServiceClient) auth.PermissionProvider {
				return permission.NewGrpcPermissionProvider(client)
			},
			handler.NewHealthHandler,
		),
		fx.Invoke(func(cfg *config.Configuration) error {
			return logging.Init(cfg.Env)
		}),
		fx.Invoke(server.NewServer),
	).Run()
}
