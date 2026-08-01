package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/velocity-center-makerspace/maintenance-tracker/db"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/config"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/pages"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/router"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/server"
)

var env *config.Environment

func main() {
	sqlDB, err := db.RunMigrations(env.DbFile)
	if err != nil {
		slog.Error("Unable to initialize database", "err", err)
	}

	qry := db.New(sqlDB)

	deps := router.Dependencies{
		DB:  sqlDB,
		Qry: qry,
		Env: env,
	}

	rt := router.New(deps)
	reg := pages.NewRegistrar()

	reg.RegisterRoutes(rt)
	rt.AddHandlers()

	ctx := context.Background()
	srv := server.NewServer(rt)

	if err := server.StartServer(ctx, srv); err != nil {
		slog.Error("Unable to start server", "err", err)
	}
}

func init() {
	logLevel := new(slog.LevelVar)
	filename := "logs/dev.log"

	if err := godotenv.Load(); err != nil {
		slog.Error("Unable to load .env file", "err", err)
	}

	env = &config.Environment{
		EnvType:        os.Getenv("SERVER_ENV"),
		DbFile:         os.Getenv("DATABASE_FILE_PATH"),
		UploadRoot:     os.Getenv("BASE_UPLOAD_DIR"),
		TempUploadRoot: os.Getenv("BASE_TEMP_DIR"),
	}

	switch env.EnvType {
	case "dev":
		filename = "logs/dev.log"
		logLevel.Set(slog.LevelDebug)
	case "prod":
		filename = "logs/prod.log"
		logLevel.Set(slog.LevelInfo)
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		slog.Error("Unable to open log file", "filepath", filename)
		return
	}

	handler := slog.NewJSONHandler(file, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(handler))

	slog.Info("Logging initialized, database file set.")
}
