package main

import (
	"fmt"
	"os"

	"github.com/gardener/etcd-druid/userInterface/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
