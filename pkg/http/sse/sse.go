package sse

import (
	"fmt"

	"github.com/cenkalti/backoff"
	provider "github.com/r3labs/sse"
)

func NewClient(uri string, events chan *provider.Event) error {
	client := provider.NewClient(uri)
	client.ReconnectStrategy = backoff.NewConstantBackOff(backoff.Stop)
	client.OnDisconnect(func(c *provider.Client) {
		fmt.Println("disconnecting")
	})

	if err := client.SubscribeChan("changelog", events); err != nil {
		return err
	}

	return nil
}
