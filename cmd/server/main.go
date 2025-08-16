package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/log"
	"github.com/joho/godotenv"
	"github.com/meeting-scheduler/internal/endpoint"
	"github.com/meeting-scheduler/internal/service"
	"github.com/meeting-scheduler/internal/transport"
	"github.com/meeting-scheduler/pkg/repository"
)

func main() {
	envFile := flag.String("env", ".env", "Environment file to load")
	flag.Parse()
	err := godotenv.Load(*envFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load %s: %v\n", *envFile, err)
		os.Exit(1)
	}
	logFilePath := "app.log"
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	var logger log.Logger
	logger = log.NewLogfmtLogger(logFile)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "root")
	dbName := getEnv("DB_NAME", "meeting_scheduler")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	repo, err := repository.NewMySQLRepository(dsn)
	if err != nil {
		logger.Log("error", err)
		os.Exit(1)
	}

	svc := service.NewService(repo)

	endpoints := endpoint.MakeEndpoints(svc)

	handler := transport.NewHTTPHandler(endpoints, logger)

	port := getEnv("PORT", "8080")

	errs := make(chan error)
	go func() {
		logger.Log("transport", "HTTP", "addr", ":"+port)
		errs <- http.ListenAndServe(":"+port, handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("exit", <-errs)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
