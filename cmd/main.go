package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/P1coFly/CarInfoEM/docs"
	"github.com/P1coFly/CarInfoEM/http-server/carinfo"
	"github.com/P1coFly/CarInfoEM/http-server/handlers/adder"
	"github.com/P1coFly/CarInfoEM/http-server/handlers/deleter"
	"github.com/P1coFly/CarInfoEM/http-server/handlers/getter"
	"github.com/P1coFly/CarInfoEM/http-server/handlers/patcher"
	"github.com/P1coFly/CarInfoEM/internal/config"
	"github.com/P1coFly/CarInfoEM/internal/storage/postgresql"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

// При изменение анотации swagger, перед запуском контейнера, надо сгенерировать документацию
// swag init -g .\cmd\main.go

// @title CarInfo App API
// @version 1.0
// @description API server for obtaining information about the car

// @host localhost:8080
// @BasePath /
func main() {
	// Загружаем файл .ENV
	if err := godotenv.Load(); err != nil {
		slog.Error("CONFIG_PATH is not set")
		os.Exit(1)
	}
	// инициализируем конфиг
	cfg := config.MustLoad()

	// инициализируем логер
	log := setupLogger(cfg.Env)

	log.Info("starting api-servies", "env", cfg.Env)
	log.Debug("cfg data", "data", cfg)

	// инициализируем storage
	storage, err := postgresql.New(cfg.HostDB, cfg.PortDB, cfg.UserDB, cfg.PasswordDB, cfg.NameDB, cfg.MigrationsPath)
	if err != nil {
		log.Error("failed to connect storage", "error", err)
		os.Exit(1)
	}
	log.Info("connect to db is successful", "host", cfg.HostDB)

	// инициализируем объект для получения информации из внешнего сервиса
	carinfo := carinfo.New(cfg.HostCarInfo)
	// инициализируем router
	router := chi.NewRouter()

	// добавляем endpoints
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	//добавляем endpoint ge
	router.Delete("/car/delete/{id}", deleter.New(log, storage))
	router.Patch("/car/patch/{id}", patcher.New(log, storage))
	router.Get("/cars", getter.New(log, storage))
	router.Post("/car/add", adder.New(log, storage, carinfo))

	//Для доступа к swagger надо пройти по URI /swagger/
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), //По URI /swagger/doc.json будет ледать спецификация в формате JSON
	))

	log.Info("starting server", slog.String("port", cfg.Port))

	// инициализируем server и запускаем
	srv := &http.Server{
		Addr:         cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server", "error", err)
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case "dev":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "prod":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
