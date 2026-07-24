package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/velocity-center-makerspace/maintenance-tracker/db"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/server"
)

var dbFile = "data/dev.db"

func main() {
	sqlDB, err := db.RunMigrations(dbFile)
	if err != nil {
		slog.Error("Unable to initialize database", "err", err)
	}

	qry := db.New(sqlDB)
	mux := server.NewMux(qry)
	// mux := server.NewMux(sqlDB)
	ctx := context.Background()
	srv := server.NewServer(mux)

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

	switch os.Getenv("ENVIRONMENT") {
	case "dev":
		filename = "logs/dev.log"
		dbFile = "data/dev.db"
		logLevel.Set(slog.LevelDebug)
	case "prod":
		filename = "logs/prod.log"
		dbFile = "data/prod.db"
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
