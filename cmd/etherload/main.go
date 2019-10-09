package main

import (
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

var rootCmd = &cobra.Command{
	Use:   "etherload",
	Short: "Load generator",
	Long:  `Load generator for SSE service`,
	Run: func(cmd *cobra.Command, args []string) {
		openedConnections := 0
		clients := make(chan int, 1)
		ticker := time.NewTicker(150 * time.Millisecond)
		events := make(chan *sse.Event)
		connectFail := false

		uri, _ := cmd.Flags().GetString("uri")

		f, err := os.Create("log.csv")
		if err != nil {
			panic("Can not create file")
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
				if !connectFail {
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

						f.Write(e.ID)
						f.Write([]byte(","))
						if len(d.Entries) > 0 {
							f.Write([]byte(d.Entries[0].PublishDate))
						} else {
							f.Write([]byte("-"))
						}

						f.Write([]byte(","))
						f.Write([]byte(time.Now().Format("2006-01-02T15:04:05.000Z0700")))
						f.Write([]byte(","))
						f.Write([]byte(strconv.Itoa(openedConnections)))
						f.Write([]byte("\n"))
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
	rootCmd.Flags().StringP("uri", "u", "", "Address to testing.")
}

var mu sync.Mutex

func main() {
	rootCmd.Execute()
}
