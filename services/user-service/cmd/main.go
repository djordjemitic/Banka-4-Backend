package main

import (
	"common/pkg/db"
	"common/pkg/logging"
	"user-service/internal/config"
	"user-service/internal/handler"
	"user-service/internal/server"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

func main() {
	fx.New(
		fx.Provide(
			config.Load,
			func(cfg *config.Configuration) (*gorm.DB, error) {
				return db.New(cfg.DB.DSN())
			},
			handler.NewHealthHandler,
		),
		fx.Invoke(func(cfg *config.Configuration) error {
			return logging.Init(cfg.Env)
		}),
		fx.Invoke(func(_ *gorm.DB) {}),
		fx.Invoke(server.NewServer),
	).Run()
}
