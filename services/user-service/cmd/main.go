package main

import (
	"common/pkg/db"
	"common/pkg/logging"
	"user-service/internal/config"
	"user-service/internal/handler"
	"user-service/internal/model"
	"user-service/internal/seed"
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
		// Seed funkcija
		fx.Invoke(func(db *gorm.DB) error {
			if err := db.AutoMigrate(&model.Employee{}, &model.Position{}); err != nil {
				return err
			}
			return seed.Run(db)
		}),
		fx.Invoke(server.NewServer),
	).Run()
}
