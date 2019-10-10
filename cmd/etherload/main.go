package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/joho/godotenv"
	"github.com/r3labs/sse"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/bialas1993/etherload/pkg/cue"
)

var buffer bytes.Buffer

var rootCmd = &cobra.Command{
	Use:   "etherload",
	Short: "Load generator",
	Long:  `Load generator for SSE service`,
	Run: func(cmd *cobra.Command, args []string) {
		uri, err := cmd.Flags().GetString("uri")
		delay, _ := cmd.Flags().GetInt("delay")
		limit, _ := cmd.Flags().GetInt("limit")

		if err != nil || len(uri) == 0 {
			panic("Uri is not")
		}

		log.Printf("Address: %s, connections limit: %d, delay new connection: %d", uri, limit, delay)

		openedConnections := 0
		clients := make(chan int, 1)
		ticker := time.NewTicker(time.Duration(delay) * time.Millisecond)
		events := make(chan *sse.Event)
		connectFail := false

		f, err := os.Create("log.csv")
		if err != nil {
			panic("Can not create log file.")
		}

		f.WriteString("id,publish,receive,connections\n")

		defer func() {
			close(clients)
			close(events)
			f.Close()
		}()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		for {
			select {
			case c := <-clients:
				log.Printf("clients: %+v\n", c)

				client := sse.NewClient(uri)
				client.ReconnectStrategy = backoff.NewConstantBackOff(backoff.Stop)
				client.OnDisconnect(func(c *sse.Client) {
					log.Println("disconnecting")
				})

				if err := client.SubscribeChan("changelog", events); err != nil {
					log.Errorf("Can not create connection, opened: %d", openedConnections)
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
					log.Debugf("event: %s\n", string(event.Data))

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
				log.Println("Closing..")
				os.Exit(0)
			}
		}
	},
}

func init() {
	log.SetLevel(log.PanicLevel)
	if strings.HasSuffix(os.Args[0], "exe/main") {
		log.SetLevel(log.DebugLevel)
	}

	godotenv.Load()
	rootCmd.Flags().StringP("uri", "u", "", "address to test")
	rootCmd.Flags().IntP("delay", "d", 150, "delay for add new connection [miliseconds]")
	rootCmd.Flags().IntP("limit", "l", 0, "connections limit (default 0)")
}

var mu sync.Mutex

func main() {
	rootCmd.Execute()
}
