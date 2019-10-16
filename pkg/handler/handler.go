package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/bialas1993/etherload/pkg/cue"
	"github.com/cenkalti/backoff"
	"github.com/r3labs/sse"
	"github.com/spf13/cobra"
)

var openedConnections int

func Load(cmd *cobra.Command, args []string) {
	uri, _ := cmd.Flags().GetString("uri")

	println("Uri")
	println(uri)
}

func load(cmd *cobra.Command, args []string) {
	uri, err := cmd.Flags().GetString("uri")
	delay, _ := cmd.Flags().GetInt("delay")
	limit, _ = cmd.Flags().GetInt("limit")

	if err != nil || len(uri) == 0 {
		panic("Uri is not")
	}

	fmt.Printf("Address: %s, connections limit: %d, delay new connection: %d\n", uri, limit, delay)

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
					
				}(event)
			}
			break
		case <-c:
			fmt.Println("Closing..")
			os.Exit(0)
		}
	}
}
