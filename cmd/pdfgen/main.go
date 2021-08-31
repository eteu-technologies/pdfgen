package main

import (
	"os"
	"os/signal"
	"sync"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	listenAddr = ":5000"
	debugMode  = true
	workdirs   sync.Map
)

func main() {
	sigCh := make(chan os.Signal, 1)
	exitCh := make(chan bool, 1)
	signal.Notify(sigCh, os.Interrupt)

	// Set up logging
	if err := configureLogging(debugMode); err != nil {
		panic(err)
	}
	defer func() { _ = zap.L().Sync() }()

	// Set up HTTP server
	router := router.New()
	srv := fasthttp.Server{
		Handler: requestLogger(router.Handler),
	}

	router.POST("/process", wrap(HandleProcess))
	router.GET("/serve/{key}/{path:*}", wrap(HandleServe))

	zap.L().Info("starting http server", zap.String("at", listenAddr))
	go func() {
		if err := srv.ListenAndServe(listenAddr); err != nil {
			zap.L().Error("failed to listen for http", zap.Error(err))
			exitCh <- true
		}
	}()

	select {
	case <-sigCh:
		zap.L().Info("got signal")
	case <-exitCh:
		// no-op
	}

	if err := srv.Shutdown(); err != nil {
		zap.L().Error("failed to shutdown the http server", zap.Error(err))
	}
}

func configureLogging(debug bool) error {
	var cfg zap.Config

	if debug {
		cfg = zap.NewDevelopmentConfig()
		cfg.Level.SetLevel(zapcore.DebugLevel)
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.Development = false
	} else {
		cfg = zap.NewProductionConfig()
		cfg.Level.SetLevel(zapcore.InfoLevel)
	}

	cfg.OutputPaths = []string{
		"stdout",
	}

	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(logger)

	return nil
}