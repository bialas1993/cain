package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/bialas1993/etherload/pkg/cue"
	"github.com/cenkalti/backoff"
	"github.com/cloudflare/cfssl/log"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/r3labs/sse"
	"github.com/spf13/cobra"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var buffer bytes.Buffer

var rootCmd = &cobra.Command{
	Use:   "etherload",
	Short: "Load generator",
	Long:  `Load generator for SSE service`,
	Run:   load,
}

func init() {
	godotenv.Load()
	rootCmd.Flags().StringP("uri", "u", "", "address to test")
	rootCmd.Flags().IntP("delay", "d", 150, "delay for add new connection [miliseconds]")
	rootCmd.Flags().IntP("limit", "l", 0, "connections limit (default 0)")
}

var mu sync.Mutex

var openedConnections = 0
var clients = make(chan int, 1)
var ticker *time.Ticker
var events = make(chan *sse.Event)
var connectFail = false
var f *os.File
var c = make(chan os.Signal, 1)
var limit = 0

func main() {
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)

		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			if scaleQuantity, err := strconv.Atoi(string(msg)); err == nil {

				if !connectFail && (openedConnections < limit || limit == 0) {
					openedConnections++

					go func(c int) {
						clients <- c
					}(openedConnections)
					continue
				}

				go func() {
					clients <- scaleQuantity
				}()
			} else {
				println("Can not parse msg.")
			}

			if err = conn.WriteMessage(msgType, msg); err != nil {
				return
			}
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "interface/http/index.html")
	})

	println("Running on http://localhost:8000/")

	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Error(err)
		}
	}()

	signal.Notify(c, os.Interrupt)

	defer func() {
		close(clients)
		close(events)
	}()

	rootCmd.Execute()

}

func load(cmd *cobra.Command, args []string) {
	uri, err := cmd.Flags().GetString("uri")
	delay, _ := cmd.Flags().GetInt("delay")
	limit, _ = cmd.Flags().GetInt("limit")

	if err != nil || len(uri) == 0 {
		panic("Uri is not")
	}

	fmt.Printf("Address: %s, connections limit: %d, delay new connection: %d\n", uri, limit, delay)

	f, err := os.Create("log.csv")
	if err != nil {
		panic("Can not create log file.")
	}

	defer func() {
		f.Close()
	}()

	f.WriteString("id,publish,receive,connections\n")

	ticker = time.NewTicker(time.Duration(delay) * time.Millisecond)

	for {
		select {
		case c := <-clients:
			fmt.Printf("clients: %+v\r", c)

			client := sse.NewClient(uri)
			client.ReconnectStrategy = backoff.NewConstantBackOff(backoff.Stop)
			client.OnDisconnect(func(c *sse.Client) {
				fmt.Println("disconnecting")
			})

			if err := client.SubscribeChan("changelog", events); err != nil {
				fmt.Printf("Can not create connection, opened: %d\n", openedConnections)
				connectFail = true
			}
			break
		case <-ticker.C:
			if !connectFail && (openedConnections < limit || limit == 0) {
				openedConnections++

				go func(c int) {
					clients <- c
				}(openedConnections)
				continue
			}

			ticker.Stop()
			break

		case event := <-events:
			if len(event.ID) > 0 {
				print('|')
				go func(e *sse.Event) {
					var d cue.Event
					json.Unmarshal(e.Data, &d)

					mu.Lock()
					defer mu.Unlock()

					buffer.Write(e.ID)
					buffer.WriteString(",")

					if len(d.Entries) > 0 {
						buffer.WriteString(d.Entries[0].PublishDate)
					} else {
						buffer.WriteString("-")
					}

					buffer.WriteString(",")
					buffer.WriteString(time.Now().Format("2006-01-02T15:04:05.000Z0700"))
					buffer.WriteString(",")
					buffer.WriteString(strconv.Itoa(openedConnections))
					buffer.WriteString("\n")

					f.Write(buffer.Bytes())
				}(event)
			}
			break
		case <-c:
			fmt.Println("Closing..")
			os.Exit(0)
		}
	}
}
