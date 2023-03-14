package main

import (
	"context"
	"fmt"
	"github.com/devtron-labs/central-api/api"
	"go.uber.org/zap"
	"net/http"
	"os"
	"time"
)

type App struct {
	MuxRouter *api.MuxRouter
	Logger    *zap.SugaredLogger
	server    *http.Server
}

func NewApp(MuxRouter *api.MuxRouter, Logger *zap.SugaredLogger) *App {
	return &App{
		MuxRouter: MuxRouter,
		Logger:    Logger,
	}
}

func (app *App) Start() {
	port := 4040 //TODO: extract from environment variable
	app.Logger.Infow("starting server on ", "port", port)
	app.MuxRouter.Init()
	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: app.MuxRouter.Router}
	app.server = server
	err := server.ListenAndServe()
	if err != nil {
		app.Logger.Errorw("error in startup", "err", err)
		os.Exit(2)
	}
}

func (app *App) Stop() {
	app.Logger.Infow("lens shutdown initiating")
	timeoutContext, _ := context.WithTimeout(context.Background(), 5*time.Second)
	app.Logger.Infow("stopping nats")

	app.Logger.Infow("closing router")
	err := app.server.Shutdown(timeoutContext)
	if err != nil {
		app.Logger.Errorw("error in mux router shutdown", "err", err)
	}
	app.Logger.Infow("closing db connection")
	app.Logger.Infow("housekeeping done. exiting now")
}
