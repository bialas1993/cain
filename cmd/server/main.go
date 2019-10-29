package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	sseclient "github.com/bialas1993/cain/pkg/http/sse"
	outLog "github.com/bialas1993/cain/pkg/logger"
	"github.com/joho/godotenv"
	"github.com/r3labs/sse"
	"github.com/sirupsen/logrus"
)

var (
	buffer            bytes.Buffer
	openedConnections int  = 0
	connectFail       bool = false
	uri               string
)

func init() {
	logrus.SetLevel(logrus.InfoLevel)

	godotenv.Load()
}

func main() {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Println("Server is starting...")

	log := outLog.New()
	clients := make(chan int)
	defer func() {
		close(clients)
		log.Close()
	}()

	flag.StringVar(&uri, "uri", "", "address to test")
	flag.Parse()

	router := http.NewServeMux()
	router.HandleFunc("/", rootHandler)
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(clients, w, r)
	})

	server := &http.Server{
		Addr:         ":3000",
		Handler:      tracing(nextRequestID)(logging(logger)(router)),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	events := make(chan *sse.Event)
	signal.Notify(quit, os.Interrupt)

	logger.Println("Server is ready to handle requests at", "https://127.0.0.1:3000")
	atomic.StoreInt32(&healthy, 1)

	go func() {
		<-quit
		logger.Println("Server is shutting down...")
		atomic.StoreInt32(&healthy, 0)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Could not listen on %s: %v\n", "https://127.0.0.1:3000", err)
			close(done)
		}
	}()

	for {
		select {
		case <-done:
			logger.Println("Server stopped")
			os.Exit(0)
			break
		case c := <-clients:
			fmt.Println("Scalling to: ", c)
			connectionsToOpen := (c - openedConnections)
			fmt.Println("To open: ", connectionsToOpen)

			if connectionsToOpen > 0 {
				for i := 0; i < connectionsToOpen; i++ {
					if connectFail != true {
						if err := sseclient.NewClient(uri, events); err != nil {
							fmt.Printf("Can not create connection, opened: %d", openedConnections-1)
							connectFail = true
						} else {
							openedConnections++
							fmt.Println("Clients: ", openedConnections)
						}
					}
				}
			}

			break
		case event := <-events:
			if len(event.ID) > 0 {
				fmt.Printf("|")

				go func(event *sse.Event, openedConnections int) {
					log.Write(&outLog.Log{event, openedConnections})
				}(event, openedConnections)
			}
			break
		}
	}
}
