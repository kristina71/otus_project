package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/kristina71/otus_project/internal/config"
	"github.com/kristina71/otus_project/internal/logger"
	"github.com/kristina71/otus_project/internal/server"
	"github.com/kristina71/otus_project/internal/services"
	"github.com/kristina71/otus_project/internal/stats"
	"github.com/kristina71/otus_project/internal/storage/sql"
	"go.uber.org/zap"
)

var TerminalSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGHUP}

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "./configs/config.yaml", "path to config file")
}

func main() {
	if err := mainImpl(); err != nil {
		log.Fatal(err)
	}
}

func mainImpl() error {
	flag.Parse()
	for _, arg := range flag.Args() {
		if arg == "version" {
			printVersion()
			return nil
		}
	}

	cnf, err := config.NewConfig(configPath)
	if err != nil {
		return fmt.Errorf("error during config reading: %w", err)
	}
	if err := logger.InitLogger(cnf.Logger); err != nil {
		return fmt.Errorf("error during logger init: %w", err)
	}
	zap.L().Info("Banner rotation service starting...")
	ctx, stop := signal.NotifyContext(context.Background(), TerminalSignals...)
	defer stop()

	zap.L().Info("rotation service storage starting...")
	dbStorage := sql.NewStorage("pgx", cnf.DB)
	if err := dbStorage.Connect(ctx); err != nil {
		return fmt.Errorf("failed to init db storage: %w", err)
	}
	defer func() {
		if err := dbStorage.Close(); err != nil {
			zap.L().Error("failed to close db storage", zap.Error(err))
		}
		zap.L().Info("rotation service storage stopped")
	}()
	zap.L().Info("rotation service storage started")

	zap.L().Info("rotation service stats publisher starting...")
	publisher, err := stats.NewPublisher(cnf.Publisher)
	if err != nil {
		return fmt.Errorf("error during stats publisher initialization: %w", err)
	}
	if err := publisher.Start(); err != nil {
		return fmt.Errorf("failed to start stats publisher: %w", err)
	}
	defer func() {
		if err := publisher.Stop(); err != nil {
			zap.L().Error("error during stats publisher stopping", zap.Error(err))
		}
		zap.L().Info("rotation service stats publisher stopped")
	}()

	app := services.NewRotationService(dbStorage, publisher)
	grpcServer := server.InitServer(app, cnf.Server)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		grpcServer.Start(stop)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		grpcServer.Stop()
	}()
	wg.Wait()
	zap.L().Info("rotation service stopped")
	return nil
}
