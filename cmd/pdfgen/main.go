package main

import (
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/eteu-technologies/pdfgen/internal/core"
)

var (
	workdirs sync.Map

	iListenAddr = "127.0.0.1:8900"
	iServerAddr = "localhost:8900"

	listenAddr                  = os.Getenv("ETEU_PDFGEN_LISTEN_ADDRESS")
	debugMode                   = strings.ToLower(os.Getenv("ETEU_PDFGEN_DEBUG")) == "true"
	noChromeSandbox             = strings.ToLower(os.Getenv("ETEU_PDFGEN_NO_CHROME_SANDBOX")) == "true"
	rendererTimeout      uint64 = 45000 // milliseconds
	maxConcurrentRenders uint64 = max(uint64(runtime.NumCPU()/2), 1)
)

func main() {
	var err error
	sigCh := make(chan os.Signal, 1)
	exitCh := make(chan bool, 1)
	signal.Notify(sigCh, os.Interrupt)

	// Set up logging
	if err := configureLogging(debugMode); err != nil {
		panic(err)
	}
	defer func() { _ = zap.L().Sync() }()

	zap.L().Info("pdfgen", zap.String("version", core.Version))

	if timeoutValue := os.Getenv("ETEU_PDFGEN_RENDERER_TIMEOUT"); timeoutValue != "" {
		if rendererTimeout, err = strconv.ParseUint(timeoutValue, 10, 64); err != nil {
			zap.L().Fatal("failed to parse renderer timeout", zap.Error(err), zap.String("value", timeoutValue))
		}
	}

	if maxConcurrentRendersValue := os.Getenv("ETEU_PDFGEN_RENDERER_MAX_CONCURRENCY"); maxConcurrentRendersValue != "" {
		if maxConcurrentRenders, err = strconv.ParseUint(maxConcurrentRendersValue, 10, 64); err != nil {
			zap.L().Fatal("failed to parse renderer max concurrency", zap.Error(err), zap.String("value", maxConcurrentRendersValue))
		}
	}

	// Set up renderer
	renderer = NewRenderer(maxConcurrentRenders)

	// Set up HTTP server
	arouter := router.New()
	asrv := fasthttp.Server{
		Handler:            requestLogger(arouter.Handler),
		MaxRequestBodySize: 100 * 1024 * 1024,
		ReadTimeout:        30 * time.Second,
		WriteTimeout:       30 * time.Second,
		IdleTimeout:        30 * time.Second,
		Logger:             zap.NewStdLog(zap.L().With(zap.String("section", "http"))),
	}

	arouter.POST("/process", wrap(HandleProcess))

	// Set up another HTTP server for internal file serving
	frouter := router.New()
	fsrv := fasthttp.Server{
		Handler:      requestLogger(frouter.Handler),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
		Logger:       zap.NewStdLog(zap.L().With(zap.String("section", "http"))),
		GetOnly:      true,
	}

	frouter.GET("/{key}/{path:*}", wrap(HandleServe))

	zap.L().Info("starting api http server", zap.String("at", listenAddr))
	go func() {
		if err := asrv.ListenAndServe(listenAddr); err != nil {
			zap.L().Error("failed to listen for api http", zap.Error(err))
			exitCh <- true
		}
	}()

	zap.L().Info("starting internal http server", zap.String("at", iListenAddr))
	go func() {
		if err := fsrv.ListenAndServe(iListenAddr); err != nil {
			zap.L().Error("failed to listen for internal http", zap.Error(err))
			exitCh <- true
		}
	}()

	select {
	case <-sigCh:
		zap.L().Info("got signal")
	case <-exitCh:
		// no-op
	}

	zap.L().Info("shutting down")

	if err := asrv.Shutdown(); err != nil {
		zap.L().Error("failed to shutdown the api http server", zap.Error(err))
	}

	if err := fsrv.Shutdown(); err != nil {
		zap.L().Error("failed to shutdown the internal http server", zap.Error(err))
	}

	if err := renderer.Close(); err != nil {
		zap.L().Error("failed to shutdown the renderer", zap.Error(err))
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

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
