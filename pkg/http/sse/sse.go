package sse

import (
	"fmt"

	"github.com/cenkalti/backoff"
	provider "github.com/r3labs/sse"
)

var openedConnections = 0

func NewClient(uri string, events chan *provider.Event) error {
	client := provider.NewClient(uri)
	client.ReconnectStrategy = backoff.NewConstantBackOff(backoff.Stop)
	client.OnDisconnect(func(c *provider.Client) {
		fmt.Println("disconnecting")
	})

	return client.SubscribeChan("changelog", events)
}
