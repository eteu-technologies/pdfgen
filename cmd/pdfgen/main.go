package main

import (
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/AubSs/fasthttplogger"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

var (
	listenAddr = ":5000"
	workdirs   sync.Map
)

func main() {
	sigCh := make(chan os.Signal, 1)
	exitCh := make(chan bool, 1)
	signal.Notify(sigCh, os.Interrupt)

	// Set up HTTP server
	router := router.New()
	srv := fasthttp.Server{
		Handler: fasthttplogger.CombinedColored(router.Handler),
	}

	router.POST("/process", wrap(HandleProcess))
	router.GET("/serve/{key}/{path:*}", wrap(HandleServe))

	go func() {
		if err := srv.ListenAndServe(listenAddr); err != nil {
			log.Println("failed to listen for http:", err)
			exitCh <- true
		}
	}()

	select {
	case <-sigCh:
		log.Println("got signal")
	case <-exitCh:
		// no-op
	}

	if err := srv.Shutdown(); err != nil {
		log.Println("failed to shutdown the http server:", err)
	}
}
