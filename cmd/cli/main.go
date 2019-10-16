package main

import (
	"fmt"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

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

func main() {
	rootCmd.Execute()
}

func load(cmd *cobra.Command, args []string) {
	uri, err := cmd.Flags().GetString("uri")
	delay, _ := cmd.Flags().GetInt("delay")
	limit, _ := cmd.Flags().GetInt("limit")

	if err != nil || len(uri) == 0 {
		panic("Uri is not defined")
	}

	fmt.Printf("Address: %s, connections limit: %d, delay new connection: %d\n", uri, limit, delay)
}
