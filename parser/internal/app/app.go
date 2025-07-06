package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/juju/zaputil/zapctx"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"parser/internal/httpadapter"
	"parser/models"
	"syscall"
	"time"
)

type app struct {
	httpAdapter httpadapter.HTTPAdapter
}

const configPath = "./config/config.json"

func (a *app) Serve(ctx context.Context) error {
	done := make(chan os.Signal, 1)
	logger := zapctx.Logger(ctx)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	logger.Info("starting server", zap.String("url", a.httpAdapter.URL()))
	go func() {
		if err := a.httpAdapter.Serve(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server error", zap.Error(err))
			log.Fatal(err.Error())
		}
	}()
	<-done

	log.Println("shutting down server")
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a.Shutdown()
	return nil
}

func (a *app) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	a.httpAdapter.Shutdown(ctx)
}

func initConfig() (*models.Config, error) {
	// Получение информации о файле
	stat, err := os.Stat(configPath)
	if err != nil {
		return nil, err
	}

	// Открытие файла
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}

	// Считывание bytes
	data := make([]byte, stat.Size())
	_, err = file.Read(data)
	if err != nil {
		return nil, err
	}

	// Десериализация в конфиг
	var config models.Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Printf("json.Unmarshal error: %v", err)
		return nil, err
	}

	log.Printf("config struct: %+v", config)

	return &config, nil
}

func New(ctx context.Context) (App, error) {
	logger := zapctx.Logger(ctx)

	logger.Info("Initializing config")
	config, err := initConfig()
	if err != nil {
		logger.Error("Config init error. %v", zap.Error(err))
		log.Fatal(err)
	}

	a := &app{httpAdapter: httpadapter.New(ctx, config)}
	return a, nil
}
