package handler

import (
	"github.com/spf13/cobra"
)

var openedConnections int

func Load(cmd *cobra.Command, args []string) {
	uri, _ := cmd.Flags().GetString("uri")

	println("Uri")
	println(uri)
}
