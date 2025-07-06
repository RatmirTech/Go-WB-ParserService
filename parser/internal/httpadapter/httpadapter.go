package httpadapter

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/juju/zaputil/zapctx"
	"net/http"
	"parser/models"
	"time"
)

type adapter struct {
	config *models.Config
	server *http.Server
}

func (a *adapter) Serve(ctx context.Context) error {
	logger := zapctx.Logger(ctx)
	logger.Info("ready to serve")
	apiRouter := chi.NewRouter()
	//apiRouter.Handle("/metrics", promhttp.Handler())

	apiRouter.Get("/", func(w http.ResponseWriter, r *http.Request) { // testing endpoint
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Pong"))
	})

	mountPath := a.config.BasePath
	if mountPath == "" {
		mountPath = "/"
	}
	apiRouter.Mount(mountPath, apiRouter)
	a.server = &http.Server{Addr: a.config.ServeAddress, Handler: apiRouter}

	// Запуск сервера в отдельной горутине
	errChan := make(chan error, 1)
	go func() {
		errChan <- a.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return a.server.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}

func (a adapter) Shutdown(ctx context.Context) {
	_ = a.server.Shutdown(ctx)
}

func (a *adapter) URL() string {
	return "http://" + a.config.ServeAddress
}

func New(ctx context.Context, config *models.Config) HTTPAdapter {
	return &adapter{
		config: config,
	}
}
